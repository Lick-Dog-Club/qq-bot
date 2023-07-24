package sysupdate

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"qq/bot"
	cfg "qq/config"
	"qq/features"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var gitCommit = ""

func init() {
	features.AddKeyword("up", "更新至最新版本", func(bot bot.Bot, content string) error {
		UpdateVersion(bot)
		return nil
	}, features.WithSysCmd(), features.WithHidden())
}

type response []struct {
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

type upBotImp interface {
	Send(string) string
}

func UpdateVersion(bot upBotImp) {
	get, _ := http.Get("https://api.github.com/repos/Lick-Dog-Club/qq-bot/commits?per_page=1")
	var data response
	defer get.Body.Close()
	json.NewDecoder(get.Body).Decode(&data)
	if gitCommit != "" && data[0].Sha[:7] != gitCommit {
		resp, _ := http.Get("https://api.github.com/repos/Lick-Dog-Club/qq-bot/actions/runs?per_page=1")
		defer resp.Body.Close()
		var runsInfo workflowRuns
		if !(len(runsInfo.WorkflowRuns) > 0 &&
			data[0].Sha == runsInfo.WorkflowRuns[0].HeadBranch &&
			runsInfo.WorkflowRuns[0].Status == "completed") {
			bot.Send("最新版本还未构建完成，请稍后～")
			return
		}

		config, err := rest.InClusterConfig()
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			return
		}
		json.NewDecoder(resp.Body).Decode(&runsInfo)
		ns := cfg.Namespace()
		pod := cfg.Pod()
		if ns != "" && pod != "" {
			bot.Send(fmt.Sprintf("更新到最新版本\n%s %s: %s\n%v", data[0].Commit.Committer.Name, data[0].Commit.Committer.Date.Local().Format("2006-01-02 15:04:05"), data[0].Commit.Message, data[0].HTMLURL))
			clientset.CoreV1().Pods(ns).Delete(context.TODO(), pod, v1.DeleteOptions{})
		}
		return
	}
	bot.Send("当前已经是最新版本~")
}

type workflowRuns struct {
	TotalCount   int `json:"total_count"`
	WorkflowRuns []struct {
		ID         int64  `json:"id"`
		HeadBranch string `json:"head_branch"`
		HeadSha    string `json:"head_sha"`
		Status     string `json:"status"`
		Conclusion string `json:"conclusion"`
	} `json:"workflow_runs"`
}
