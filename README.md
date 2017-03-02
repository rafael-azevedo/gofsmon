# **Gofsmon** <sup><sub>_Automated File Cleanup In GO_</sub></sup>
----

1. [Overview](#overview)
2. [Getting Started](#getting-started)
  * [Creating Config File](#creating-config-file)
  * [Setting Enviroment](#setting-enviroment)
  * [Installation](#installation)
  * [Running Gofsmon](#running-gofsmonnap)


## Overview

**Gofsmon** reads from a config file and cleans files accordsing to certain rules. 

Currently runs on linux and osx

## **Getting Started**

### Creating Config File

Currently gofsmon reads from a YAML file.

There are two objecst that can be created. Each object can take similar parameters with the exception of the threshold and time parameters. For each file regex you would like to clean you need a new object

#### object

* timefs - These objects clean directories for any files that match the regex and older than x seconds
* thresholdfs - These objects clean directories for any files that match the regex and whose filesystem mount point exceeds threshold x

#### parameters

* mountpoint - This is the mountpoint that is being checked
* truncate - Bool can be provided with yes or no. If set to yes truncate removes any file matching the regex except the latest and truncates the file to an empty file 
* log
    * dir - This is the directory where the logs are located
    * regex - This is the regex we are checking in the above directory
* time - Time in seconds. Can only be used with timefs object. This says clean all files in the directory that exceeds x seconds
* threshold - % that filesystem must exceed in order to trigger cleaning. This says clean all files in the directory if the directory is x% full or more
* script - Script to be executed if trigger conitions are met

for example `config.yaml` with following content:
```yaml
---
  timefs :
    - mountpoint: '/'
    log: 
        dir: '/.dev/log/'
        regex: 'test*.log'
    time: 10800
    truncate: no
     - mountpoint: '/'
     log: 
        dir: '/.dev/log/'
        regex: 'process*.log'
    time: 26280
    truncate: no
    - mountpoint: '/'
    log: 
        dir: '/.dev/log/'
        regex: 'somelog*.log'
    time: 10
    truncate: yes
  thresholdfs:
    - mountpoint: '/'
    log:
        dir: '/.dev/log/'
        regex: 'another*.log'
    threshold: 10
    truncate: yes
    - mountpoint: '/'
    log:
        dir: '/.dev/log/'
        regex: 'process*.log'
    threshold: 50
    truncate: no 
    script: '/.dev/scripts/clean.sh'
```

### Setting Enviroment
The config file location should be set to **GOFSMONCONF** in the enviroment

either export log location to your current enviroment 
```
$ export GOFSMONCONF=~/.go/src/github.com/rafael-azevedo/gofsmon/config.yaml
```
or set **GOFSMONCONF** in your path  

### Installation 
The build script can be found at ~/.go/src/github.com/rafael-azevedo/gofsmon/scripts/build_gofsmon.sh

```
$ cd github.com/Staples-Inc/gofsmon/scripts 
$ ./build_gofsmon.sh
2016-10-18 16:03:14 UTC [info] project path: ~/.go/src/github.com/rafael-azevedo/gofsmon
2016-10-18 16:03:14 UTC [info] Building gofsmon to ~/.go/src/github.com/rafael-azevedo/gofsmon/bin/
```

### Running Gofsmon
```
$ cd github.com/rafael-azevedo/gofsmon/bin
$ ./gofsmon
```
