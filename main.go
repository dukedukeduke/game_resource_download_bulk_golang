package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
)


var (
	urlPrefix string = "http://[mainapp].domain.com/"

	allSubAppList map[string]map[string]string = map[string]map[string]string{
		"subapp0001": map[string]string{
			// ios-android
			"prelaunch": "0-0",
      },

		"subapp0002": map[string]string{
			// ios-android
			"prelaunch": "0-0",
      }}

	mainappSubappResourceType map[string][]string = map[string][]string{
		"subapp0001": []string{
			".txt", ".zip", ".zip.txt", "lua.zip", "lua.zip.txt", "raw.zip",
			"raw.zip.txt"},

		"subapp0002": []string{
			".txt", ".zip", ".zip.txt", "lua.zip", "lua.zip.txt", "raw.zip",
			"raw.zip.txt"},

	}

	mainappBucket map[string]string = map[string]string{
		"subapp0001": "food0001",
		"subapp0002": "food0002",
	}

	subappNameResourceSpecial map[string]map[string]string =
		map[string]map[string]string{
			"subapp0001" : map[string]string{"prelaunch": "mainapp"},
			"subapp0002" : map[string]string{"prelaunch": "mainapp"},
	}
)

func downloadOnlySubapp(mainapp, subapp, stage, platform,
	currentDir string, wg *sync.WaitGroup) error{
		var(
			version string
			resourceType string
			subApp string
			url string
			r *http.Response
			err error
			body []byte
			dataPath string
			filePath string
			f *os.File
		)
		if wg != nil{
			defer wg.Done()
		}
		if _, ok := allSubAppList[mainapp]; !ok{
			fmt.Println("Mainapp not exists:", mainapp)
			return nil
		}
		if _, ok := allSubAppList[mainapp][subapp]; !ok{
			fmt.Println("Subapp not exists:", subapp)
			return nil
		}
		version = strings.Split(allSubAppList[mainapp][subapp], "-")[0]
		for _, resourceType = range mainappSubappResourceType[mainapp]{
			if _, ok := subappNameResourceSpecial[mainapp]; !ok{
				fmt.Println("Mainapp not exists:", mainapp)
				return nil
			}else{
				if _, ok = subappNameResourceSpecial[mainapp][subapp];
					ok{
					subApp = subappNameResourceSpecial[mainapp][subapp]
				}else{
					subApp = subapp
				}
			}
			url = strings.Replace(urlPrefix, "[mainapp]", mainapp, -1) +
				strings.Join([]string{platform, stage, version,
				subApp + resourceType}, "/")
			if r, err = http.Get(url); err != nil{
				fmt.Printf("error happened for download url:%v", url)
			}
			defer r.Body.Close()
			fmt.Println(r.StatusCode)
			if r.StatusCode == 200{
				if body, err = ioutil.ReadAll(r.Body);err != nil{
					fmt.Println("读取输出错误， URL：", url)
					return nil
				}else{
					dataPath = path.Join(currentDir, "data", mainapp, platform,
						stage, version)
					if _, err = os.Stat(dataPath); os.IsNotExist(err){
						fmt.Println("文件夹不存在：",dataPath)
						if err = os.MkdirAll(dataPath, 0777); err != nil{
							fmt.Println("创建文件夹失败：",
								strings.Join([]string{mainapp, platform,stage,
								version}, "/"))
							fmt.Printf(err.Error())
							return nil
						}
						fmt.Println("创建文件夹成功：", dataPath)
					}
					filePath = path.Join(dataPath, subApp + resourceType)
					if f, err = os.OpenFile(filePath,
						os.O_WRONLY | os.O_CREATE| os.O_APPEND,
						0777); err != nil{
						fmt.Println("打开文件错误：", filePath)
						break
					}else{
						if _, err = f.Write(body); err != nil{
							fmt.Println("写入文件错误：", filePath)
							break
						}else{
							fmt.Println("写入文件成功：", filePath)
						}
					}
				}
			}else{
				fmt.Println("Error happened when download data for url:",
					url)
			}
		}
		return nil
}

func main(){
	var (
		mainapp *string
		subapp *string
		mode *string
		platform *string
		stage *string
		currentDir string
		err error
		_subapp string
	)
	if currentDir, err = os.Getwd();err != nil{
		panic("Error happend: get current dir error")
	}
	mainapp = flag.String("mainapp", "", "the bucket you " +
		"want to get files from")
	subapp = flag.String("subapp", "", "the subapp you " +
		"want to get files")
	// if mode == "all", means get all subapp listed above info
	mode = flag.String("mode", "", "the mode you " +
		"want to get files")
	stage = flag.String("stage", "", "the stage you " +
		"want to get files")
	platform = flag.String("platform", "", "the platform " +
		"you want to get files")
	flag.Parse()

	if *platform != "ios"{
		fmt.Println("Current platform not support:", *platform)
		os.Exit(1)
	}

	if *stage != "test" && *stage != "production"{
		fmt.Println("Current stage not support:", *stage)
		os.Exit(1)
	}

	if *subapp != "" && *mode != "all"{
		// get one subapp files
		downloadOnlySubapp(*mainapp, *subapp, *stage, *platform, currentDir, nil)
	}else if *subapp == "" && *mode == "all" && *mainapp != ""{
		// get all subapp files for mainapp
		var (
			wg sync.WaitGroup
		)
		if _, ok := allSubAppList[*mainapp]; !ok{
			fmt.Println("Mainapp not exists:", *mainapp)
			os.Exit(1)
		}
		for _subapp, _ = range allSubAppList[*mainapp]{
			wg.Add(1)
			go downloadOnlySubapp(*mainapp, _subapp, *stage, *platform,
				currentDir, &wg)
		}
		wg.Wait()
	}else{
		fmt.Println("Error happened because of parameters error")
		os.Exit(1)
	}
}
