package main

type Downloader interface {
	CreateMultiDownloadTasks(urls []string)
}
