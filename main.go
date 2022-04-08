package main

import (
	"fmt"
	planfix2 "github.com/popstas/planfix-go/planfix"
	"github.com/tj/go-dropbox"
	"log"
	"net/http"
	"regexp"
	"screenshotmonitor-to-planfix/bitly"
	"screenshotmonitor-to-planfix/config"
	"screenshotmonitor-to-planfix/sm"
	"strconv"
	"strings"
	"time"
)

func main() {
	bApi := bitly.New(config.GetConfig().Bitly.Token)
	smApi := sm.New(config.GetConfig().ScreenshotMonitor.Token)
	db := dropbox.New(dropbox.NewConfig(config.GetConfig().Dropbox.Token))
	userRelations := getUsers()

	pfConfig := config.GetConfig().Planfix
	planfix := planfix2.New(
		"https://api.planfix.ru/xml/",
		pfConfig.ApiKey,
		pfConfig.Account,
		pfConfig.User,
		pfConfig.Password,
	)

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	yesterday = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, time.Local)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).Add(-1 * time.Second)

	for smId, _ := range userRelations {

		log.Println("Activities of user #", smId)

		actResp, err := smApi.GetActivities(smApi.NewRequestActivityForUser(smId, yesterday))

		if err != nil {
			printError(err)
			continue
		}

		for _, activity := range *actResp {
			log.Println("Activity", activity.Note)
			log.Println("Started", activity.FromFormatted(), activity.FromFormattedTime())
			log.Println("Finished", activity.ToFormatted(), activity.ToFormattedTime())

			if activity.GetFrom().Before(yesterday) || activity.GetTo().After(today) {
				log.Println("Wrong time range. Skipping activity.", "\n", "--------------")
				continue
			}

			getRefId(activity)
			if activity.PlanfixId == 0 {
				log.Println("not fount planfix task id in activity `note` field. skipping activity.")
				continue
			}

			dirPath := SanitizeFilename(
				`/` +
					activity.GetFrom().Format("2006-01-02") +
					`/` +
					activity.Note +
					`/` +
					activity.FromFormattedTime() +
					" - " +
					activity.ToFormattedTime(),
			)

			if uploadScreenshotsToDropbox(getScreenshots(smApi, activity), dirPath, db) > 0 {
				shareUrl := getDropboxFolderPath(db, dirPath)
				activity.ScreenshotUrl = shareUrl

				log.Println("Shortening link")
				if shortenResp, err := bApi.Shorten(bitly.Url(shareUrl)); err != nil {
					activity.ScreenshotUrl = shortenResp.Link
					log.Println("Short URL:", activity.ScreenshotUrl)
				}

				if activity.ScreenshotUrl != "" {
					log.Println("Converting URL to HTML-link")

					activity.ScreenshotUrl = fmt.Sprintf("<a href=%q>%s</a>", activity.ScreenshotUrl, activity.ScreenshotUrl)
				}
			}

			log.Println("Fetching planfix task info")
			task, err := planfix.TaskGet(activity.PlanfixId, activity.PlanfixId)

			if err != nil {
				printError(err)
				continue
			}

			activity.PlanfixRealId = task.Task.ID

			var notifiedList []planfix2.XMLResponseUser

			notifiedList = append(notifiedList, planfix2.XMLResponseUser{ID: task.Task.OwnerId})
			for _, user := range task.Task.WorkersUsers.Users {
				notifiedList = append(notifiedList, planfix2.XMLResponseUser{ID: user.ID})
			}

			var splitActivities []*sm.GetActivityResponseItem

			splitActivities = append(splitActivities, activity)

			if activity.GetFrom().Day() != activity.GetTo().Day() {
				log.Println("Splitting activity because it ended in another day")
				newActivity := *activity
				newActivity.ScreenshotUrl = ""

				newActivity.From = time.Date(activity.GetTo().Year(), activity.GetTo().Month(), activity.GetTo().Day(), 0, 0, 0, 0, time.Local).Unix()

				log.Println("new activity date:", newActivity.FromFormatted(), newActivity.FromFormattedTime(), " - ", newActivity.ToFormattedTime())
				splitActivities = append(splitActivities, &newActivity)

				activity.To = time.Date(activity.GetFrom().Year(), activity.GetFrom().Month(), activity.GetFrom().Day(), 23, 59, 59, 999, time.Local).Unix()
				log.Println("old activity date:", activity.FromFormatted(), activity.FromFormattedTime(), " - ", activity.ToFormattedTime())
			}

			for _, splitActivity := range splitActivities {
				sendToPlanfix(splitActivity, &planfix, &notifiedList)
			}
		}
	}

}

func sendToPlanfix(activity *sm.GetActivityResponseItem, planfix *planfix2.API, notifiedList *[]planfix2.XMLResponseUser) {
	fields := []planfix2.XMLRequestAnaliticField{
		{
			FieldID: 3,
			Value:   activity.FromFormatted(),
		},
		{
			FieldID: 4,
			Value: planfix2.XMLRequestAnaliticTimePeriodValue{
				Begin: activity.FromFormattedTime() + ":00",
				End:   activity.ToFormattedTime() + ":00",
			},
		},
		{
			FieldID: 5,
			Value:   39,
		},
		{
			FieldID: 4045,
			Value: struct {
				ID int `xml:"id"`
			}{getUsers()[activity.EmploymentId]},
		},
		{
			FieldID: 4048,
			Value:   activity.ScreenshotUrl,
		},
	}
	analytics := []planfix2.XMLRequestActionAnalitic{
		{ID: 3, ItemData: fields},
	}

	log.Println("Sending planfix action with analytics")
	_, err := planfix.ActionAdd(planfix2.XMLRequestActionAdd{
		Description:  activity.Note,
		TaskID:       activity.PlanfixRealId,
		IsHidden:     0,
		Analitics:    analytics,
		NotifiedList: *notifiedList,
	})

	if err != nil {
		printError(err)
	}
}

func getDropboxFolderPath(db *dropbox.Client, dirPath string) string {
	log.Println("Getting dropbox folder share link")
	links, err := db.Sharing.ListSharedLinks(&dropbox.ListShareLinksInput{Path: dirPath})
	if err != nil {
		printError(err)
		//continue
	}

	var shareUrl string
	if len(links.Links) > 0 {
		log.Println("Path already shared")
		shareUrl = links.Links[0].URL
	} else {
		log.Println("Path wasn't shared previously")

		sharedLink, err := db.Sharing.CreateSharedLink(&dropbox.CreateSharedLinkInput{Path: dirPath})

		if err != nil {
			printError(err)
		}

		shareUrl = sharedLink.URL
	}

	log.Println("Share URL:", shareUrl)
	return shareUrl
}

//uploadScreenshotsToDropbox upload screenshots from SM to Dropbox and retrieve share link to folder
func uploadScreenshotsToDropbox(screenshots *sm.GetScreenshotsResponse, dirPath string, db *dropbox.Client) int {
	activeScreenshots := 0

	for _, screenshot := range *screenshots {
		log.Println("Screenshot URL:", screenshot.Url)
		if strings.Contains(screenshot.Url, "unlock") {
			log.Println("Wrong screenshot URL. Skipping!")
			continue
		}

		filePath := dirPath + `/` + strconv.Itoa(screenshot.Id) + ".jpg"

		log.Println("Downloading... ")
		sFile, _ := http.Get(screenshot.Url)

		log.Println("Uploading to Dropbox at ", filePath)
		_, err := db.Files.Upload(&dropbox.UploadInput{
			Path:       filePath,
			Mode:       dropbox.WriteModeOverwrite,
			AutoRename: true,
			Mute:       true,
			Reader:     sFile.Body,
		})

		if err != nil {
			printError(err)
			continue
		}
		sFile.Body.Close()

		activeScreenshots++
	}
	return activeScreenshots
}

func getScreenshots(smApi *sm.SM, activity *sm.GetActivityResponseItem) *sm.GetScreenshotsResponse {
	log.Println("Getting screenshots")
	screenshotReq := smApi.NewRequestScreenshotsArgs(activity.Id)
	screenshotsResp, err := smApi.GetScreenshots(screenshotReq)

	if err != nil {
		printError(err)
	}

	log.Println("got", len(*screenshotsResp), "screenshots")
	return screenshotsResp
}

func getRefId(activity *sm.GetActivityResponseItem) {
	re := regexp.MustCompile(`#([\d]+)\s`)
	matches := re.FindSubmatch([]byte(activity.Note))
	if matches != nil {
		pfId := string(matches[1])
		activity.PlanfixId, _ = strconv.Atoi(pfId)

		activity.Note = strings.TrimSpace(strings.ReplaceAll(activity.Note, "#"+pfId, ""))
	}
}

func printError(err error) {
	if err != nil {
		log.Println(err)
	}
}

func getUsers() map[int]int {
	return config.GetConfig().Users
}
