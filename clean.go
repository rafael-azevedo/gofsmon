package gofsmon

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"time"

	yaml "gopkg.in/yaml.v2"
)

//CleanService is any type that can execute the CleanDir function
type CleanService interface {
	CleanDir() error
}

//MCleanService allows us to range over types of CleanService interface
type MCleanService []CleanService

//CleanDir cleans with appropriate rules for any of the items that satisfies the CleanService interface
func (mc MCleanService) CleanDir() error {
	for _, object := range mc {
		err := object.CleanDir()
		if err != nil {
			return err
		}
	}
	return nil
}

//NewTFS returns a MConfig that satsifies a cleanservice interface from a []byte
func (mc *MCleanService) NewTFS(config []byte) error {
	var fs filesystem

	if err := yaml.Unmarshal(config, &fs); err != nil {
		return err
	}

	//for TimeFileSystem  Get information for all files in log directory matching regex and add it to the slice of cleanservice interface
	for _, item := range fs.Timefs {
		item.Log.setLogInfo()
		*mc = append(*mc, item)
	}

	//For ThresholdFileSystem Get information for all files in the log directories if the threshold of the mountpoint is exceeded
	for _, item := range fs.Thresholdfs {
		if item.getPercentUsed() > item.Threshold {
			item.Log.setLogInfo()
			*mc = append(*mc, item)
		}
	}

	return nil
}

//TimeFileSystem is a type that stores info needed to clean by time (cleaning any file regex that exceeds X seconds)
type TimeFileSystem struct {
	MountPoint string   `yaml:"mountpoint"`
	Log        LogRegex `yaml:"log"`
	Time       int      `yaml:"time"`
	Truncate   bool     `yaml:"truncate"`
}

//CleanDir Cleans any matching files that are older then x seconds
func (fs TimeFileSystem) CleanDir() error {
	log.Println("Cleaning by time")
	log.Printf("%+v\n", fs)

	for i := range fs.Log.Finfo {
		fileName := fs.Log.Dir + fs.Log.Finfo[i].Name()
		fileModTime := fs.Log.Finfo[i].ModTime()
		timeSinceModTime := time.Since(fileModTime)
		timeToDelete := time.Duration(fs.Time) * time.Second

		switch {
		case timeSinceModTime > timeToDelete:
			switch {
			case i == 0:
				if fs.Truncate == true {
					log.Println("Truncating", fileName, timeSinceModTime, timeToDelete)
					err := os.Truncate(fileName, 0)
					if err != nil {
						return err
					}
				} else {
					log.Println("deleting", fileName, timeSinceModTime, timeToDelete)
					err := os.Remove(fileName)
					if err != nil {
						return err
					}
				}
			case i > 0:
				log.Println("deleting", fileName, timeSinceModTime, timeToDelete)
				err := os.Remove(fileName)
				if err != nil {
					return err
				}
			}
		}

	}
	return nil
}

//ThresholdFileSystem is a type that stores info needed to clean by threshold
type ThresholdFileSystem struct {
	MountPoint string   `yaml:"mountpoint"`
	Log        LogRegex `yaml:"log"`
	Threshold  float64  `yaml:"threshold"`
	Truncate   bool     `yaml:"truncate"`
}

//CleanDir Threshold FileSystem Cleans matching files from filesystem when static threshold is exceeded
func (fs ThresholdFileSystem) CleanDir() error {
	log.Println("Cleaning by threshold")
	log.Printf("%+v\n", fs)

	for i := range fs.Log.Finfo {
		switch {
		case i == 0:
			if fs.Truncate == true {
				fileName := fs.Log.Dir + fs.Log.Finfo[i].Name()
				log.Println("Truncating", fileName)
				err := os.Truncate(fileName, 0)
				if err != nil {
					return err
				}
			}
		case i > 0:
			fileName := fs.Log.Dir + fs.Log.Finfo[i].Name()
			log.Println("deleting", fileName)
			err := os.Remove(fileName)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type filesystem struct {
	Timefs      []TimeFileSystem      `yaml:"timefs"`
	Thresholdfs []ThresholdFileSystem `yaml:"thresholdfs"`
}

//LogRegex holds directory and slice containing log regex and details
type LogRegex struct {
	Dir   string        `yaml:"dir"`
	Regex string        `yaml:"regex"`
	Finfo []os.FileInfo `yaml:"loginfo,omitempty"`
}

//ReadYamal reads in config from yaml file and return is as a []byte
func ReadYamal(configfile string) ([]byte, error) {
	source, err := ioutil.ReadFile(configfile)
	if err != nil {
		return nil, err
	}
	return source, nil
}

//GetPercentUsed returns the amount of space used in % for the ThresholdFileSystem
func (fs ThresholdFileSystem) getPercentUsed() float64 {

	buf := syscall.Statfs_t{}
	err := syscall.Statfs(fs.MountPoint, &buf)
	if err != nil {
		log.Fatal(err)
	}

	totalBytes := uint64(buf.Bsize) * buf.Blocks
	userFreeBytes := uint64(buf.Bsize) * buf.Bavail
	rootFreeBytes := uint64(buf.Bsize)*buf.Bfree - userFreeBytes
	usedBytes := totalBytes - userFreeBytes - rootFreeBytes
	usedBytesPercent := (100 * float64(usedBytes) / float64(totalBytes))

	return usedBytesPercent
}

//setLogInfo returns os.Fileinfo for all files matching LogRegex
func (l *LogRegex) setLogInfo() {

	globPath := l.Dir + l.Regex
	files, err := filepath.Glob(globPath)
	if err != nil {
		log.Println(err)
	}

	for _, file := range files {
		stat, err := os.Stat(file)
		if err != nil {
			log.Println(err)
		}
		l.Finfo = append(l.Finfo, stat)

	}
}

//functions required to implement sort by time on os.Fileinfo
type byNewest []os.FileInfo

func (a byNewest) Len() int           { return len(a) }
func (a byNewest) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byNewest) Less(i, j int) bool { return a[i].ModTime().After(a[j].ModTime()) }
