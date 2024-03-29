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
	}, features.WithSysCmd(), features.WithHidden(), features.WithGroup("system"))
	features.AddKeyword("restart", "重启服务", func(bot bot.Bot, content string) error {
		if b, f := restart(); b {
			bot.Send("系统现在重启")
			f()
		}
		return nil
	}, features.WithSysCmd(), features.WithHidden(), features.WithGroup("system"))
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

type ghError struct {
	Message string `json:"message"`
}

func UpdateVersion(bot upBotImp) {
	get, err := http.Get("https://api.github.com/repos/Lick-Dog-Club/qq-bot/commits?per_page=1")
	if err != nil {
		bot.Send(err.Error())
		return
	}
	var data response
	defer get.Body.Close()
	if get.StatusCode >= 400 {
		var ghErr ghError
		json.NewDecoder(get.Body).Decode(&ghErr)
		bot.Send(ghErr.Message)
		return
	}
	if err := json.NewDecoder(get.Body).Decode(&data); err != nil {
		bot.Send(err.Error())
		return
	}
	if gitCommit != "" && data[0].Sha[:7] != gitCommit {
		resp, _ := http.Get("https://api.github.com/repos/Lick-Dog-Club/qq-bot/actions/runs?per_page=1")
		defer resp.Body.Close()
		var runsInfo workflowRuns
		json.NewDecoder(resp.Body).Decode(&runsInfo)
		if !(len(runsInfo.WorkflowRuns) > 0 &&
			data[0].Sha == runsInfo.WorkflowRuns[0].HeadSha &&
			runsInfo.WorkflowRuns[0].Status == "completed") {
			bot.Send("最新版本还未构建完成，请稍后～")
			return
		}

		if b, f := restart(); b {
			bot.Send(fmt.Sprintf("更新到最新版本\n%s %s: %s\n%v", data[0].Commit.Committer.Name, data[0].Commit.Committer.Date.Local().Format("2006-01-02 15:04:05"), data[0].Commit.Message, data[0].HTMLURL))
			f()
		}
		return
	}
	bot.Send("当前已经是最新版本~")
}

func restart() (bool, func()) {
	config, _ := rest.InClusterConfig()
	clientset, _ := kubernetes.NewForConfig(config)
	ns := cfg.Namespace()
	pod := cfg.Pod()
	if ns != "" && pod != "" {
		return true, func() {
			clientset.CoreV1().Pods(ns).Delete(context.TODO(), pod, v1.DeleteOptions{})
		}
	}
	return false, func() {}
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
