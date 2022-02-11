package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type Watch struct {
	watch *fsnotify.Watcher
}

func (w *Watch) watchDir(dir string) {
	// 1 检查提供的是否是 directory，如果不是则退出

	// 2. filepath.Walk 遍历该目录中的每个文件（该文件可能是普通文件或者目录）
	//		Walk 会在每次遍历到文件时调用 WalkFunc 函数
	//		func filepath.Walk(root string, fn WalkFunc) error
	//		type WalkFunc func(path string, info os.FileInfo, err error) error
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			// 如果是目录就获取该目录的绝对路径
			path, err := filepath.Abs(path)
			if err != nil {
				return err
			}

			// 添加目录监控
			err = w.watch.Add(path)
			if err != nil {
				return err
			}
		}
		return nil
	})

	// 3.开始监控
	log.Println("监控服务已经启动")
	go func() {
		// 循环监控文件变动
		// 文件变化事件(event)会写入 w.watch.Events 管道中，
		// 我们只需要 for-select 循环从管道中读取 event 即可
		for {
			select {
			// 从 w.watch.Events 中读取事件
			case event := <-w.watch.Events:
				switch {
				// Create event
				case event.Op&fsnotify.Create == fsnotify.Create:
					_, fileName := filepath.Split(event.Name) // 获取文件的相对路径
					log.Println(event.Op.String(), fileName)
					file, err := os.Stat(event.Name)
					if err == nil && file.IsDir() { // 如果创建的是目录，就增加该目录的监控
						w.watch.Add(event.Name)
						log.Println("增加监控:", fileName)
					}
				// Remove event
				case event.Op&fsnotify.Remove == fsnotify.Remove:
					_, fileName := filepath.Split(event.Name) // 获取文件的相对路径
					log.Println(event.Op.String(), fileName)
					file, err := os.Stat(event.Name)
					if err == nil && file.IsDir() { // 如果删除的是目录，就删除该目录的监控
						w.watch.Remove(event.Name)
						log.Println("删除监控:", fileName)
					}
				// Write event
				case event.Op&fsnotify.Write == fsnotify.Write:
					_, fileName := filepath.Split(event.Name)
					log.Println(event.Op.String(), fileName)
				// Rename event
				case event.Op&fsnotify.Rename == fsnotify.Rename:
					_, fileName := filepath.Split(event.Name) // 获取文件的相对路径
					log.Println(event.Op.String(), fileName)
				// Chmod event
				case event.Op&fsnotify.Chmod == fsnotify.Chmod:
					_, fileName := filepath.Split(event.Name)
					log.Println(event.Op.String(), fileName)
				}

			// 从 w.watch.Errors 管道中获取错误
			case err := <-w.watch.Errors:
				log.Println("error:", err)
				return
			}
		}
	}()
}

func main() {
	dirPath := "./dir"
	watch, _ := fsnotify.NewWatcher()
	w := Watch{
		watch: watch,
	}
	w.watchDir(dirPath)
	select {}
}
