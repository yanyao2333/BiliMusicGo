package main

import (
	"io"
	"net/http"
	"os"
	"sync"
)

// Task 结构体表示一个下载任务
type Task struct {
	URL            string
	FileName       string
	FavourFolderId int
	DstDir         string // 下载后的文件由于还需要转码 添加元数据, 所以先存在./tmp文件夹下,这个字段定义了最终会转移到的目标文件夹
}

// Result 结构体表示一个下载结果
type Result struct {
	Task   Task
	Err    error
	Status string
}

// Downloader 结构体管理下载过程
type Downloader struct {
	concurrency int
	tasks       chan Task
	results     chan Result
}

// NewDownloader 创建一个新的 Downloader 实例
func NewDownloader(concurrency int) *Downloader {
	return &Downloader{
		concurrency: concurrency,
		tasks:       make(chan Task),
		results:     make(chan Result),
	}
}

// Start 启动下载器
func (d *Downloader) Start() {
	var wg sync.WaitGroup
	for i := 0; i < d.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case task, ok := <-d.tasks:
					if !ok {
						return
					}
					d.results <- d.downloadFile(task)
				}
			}
		}()
	}
	wg.Wait()
}

// AddTask 添加一个下载任务
func (d *Downloader) AddTask(task Task) {
	d.tasks <- task
}

// Stop 停止下载器
func (d *Downloader) Stop() {
	close(d.tasks)
	close(d.results)
}

// downloadFile 下载单个文件
func (d *Downloader) downloadFile(task Task) Result {
	resp, err := http.Get(task.URL)
	if err != nil {
		return Result{Task: task, Err: err, Status: "Failed"}
	}
	defer resp.Body.Close()

	out, err := os.Create(task.FileName)
	if err != nil {
		return Result{Task: task, Err: err, Status: "Failed"}
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return Result{Task: task, Err: err, Status: "Failed"}
	}

	return Result{Task: task, Err: nil, Status: "Success"}
}
