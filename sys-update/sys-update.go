package sys_update

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"qq/bot"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var gitCommit = ""

type Response []struct {
	Sha    string `json:"sha"`
	NodeID string `json:"node_id"`
	Commit struct {
		Author struct {
			Name  string    `json:"name"`
			Email string    `json:"email"`
			Date  time.Time `json:"date"`
		} `json:"author"`
		Committer struct {
			Name  string    `json:"name"`
			Email string    `json:"email"`
			Date  time.Time `json:"date"`
		} `json:"committer"`
		Message      string `json:"message"`
		URL          string `json:"url"`
		CommentCount int    `json:"comment_count"`
		Verification struct {
			Verified  bool        `json:"verified"`
			Reason    string      `json:"reason"`
			Signature interface{} `json:"signature"`
			Payload   interface{} `json:"payload"`
		} `json:"verification"`
	} `json:"commit"`
	URL         string `json:"url"`
	HTMLURL     string `json:"html_url"`
	CommentsURL string `json:"comments_url"`
}

func Version() string {
	return gitCommit
}

func UpdateVersion(msg bot.Message) {
	get, _ := http.Get("https://api.github.com/repos/Lick-Dog-Club/qq-bot/commits?per_page=1")
	var data Response
	defer get.Body.Close()
	json.NewDecoder(get.Body).Decode(&data)
	log.Println(data[0].Sha[:7])
	if gitCommit != "" && data[0].Sha[:7] != gitCommit {
		config, err := rest.InClusterConfig()
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			return
		}
		ns := os.Getenv("APP_NAMESPACE")
		pod := os.Getenv("POD_NAME")
		if ns != "" && pod != "" {
			bot.Send(msg, fmt.Sprintf("更新到最新版本 [%s %s: %s](%v)...", data[0].Commit.Committer.Name, data[0].Commit.Committer.Date.Format("2006-01-02 15:04:05"), data[0].Commit.Message, data[0].HTMLURL))
			clientset.CoreV1().Pods(ns).Delete(context.TODO(), pod, v1.DeleteOptions{})
		}
		return
	}
	bot.Send(msg, "当前已经是最新版本~")
}
