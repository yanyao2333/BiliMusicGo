package main

type Downloader interface {
	CreateDownloadTask(url string) error
}
