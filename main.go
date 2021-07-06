package main

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
)
//todo:将配置项导出为配置文件，解析后使用
//读取模式，默认为0，直接读取二进制文件,当值为1时，使用ffmpeg进行读取
var readMod=0
//放视频文件的文件目录
var basePath string

//开发模式下，屏蔽命令行参数，改为内置写死
var debugMod=false

//读取解析后的文件时长数据
func loadData(infoPath string)map[string]int{
	var data=make(map[string]int)
	content, err := ioutil.ReadFile(infoPath)
	if err != nil {
		fmt.Println("读取文件失败！err:", err)
		return nil
	}
	contentLines:=strings.Split(string(content),"\n")
	for _,line:=range contentLines{
		if line==""{
			continue
		}
		item:=strings.Split(line,"\t")
		value,_:=strconv.Atoi(item[1])
		data[item[0]]=value
	}
	return data
}

//保存文件时长数据
func saveData(data map[string]int,infoPath string){
	str:=""
	for filename,duration:=range data{
		str+=fmt.Sprintf("%s\t%d\n",filename,duration)
	}
	err := ioutil.WriteFile(infoPath, []byte(str), 0666)
	if err != nil {
		fmt.Println("保存info数据失败！, err:", err)
		return
	}
}

//获取所有文件时长
func getInfo(data map[string]int,filePathList []string){
	for index,filePath:=range filePathList{
		fmt.Printf("进度%d/%d,解析 %s 的数据\n",index,len(filePathList),filePath)
		fileName := getFileName(filePath)
		if readMod==0{
			data[fileName] = getDurationByReadFile(filePath)
		}else{

			data[fileName] = getDurationByFfmpeg(filePath)
		}
	}
}


func getFileName(filePath string) string {
	fileNameP := strings.LastIndex(filePath, "\\")
	fileName := filePath[fileNameP+1:]
	return fileName
}

func getFilePath(Path string) []string {
	Files, _ := ioutil.ReadDir(Path)
	var FilePathList []string
	for _, f := range Files {
		if f.Name()=="pass"||f.Name()=="info.txt"{
			continue
		}
		FilePathList = append(FilePathList, fmt.Sprintf("%s\\%s", Path, f.Name()))
	}
	return FilePathList
}

//格式化数据转为秒
func translate2Second(duration string)(totalSecond int){
	hour,_:=strconv.Atoi(duration[:2])
	minute,_:=strconv.Atoi(duration[3:5])
	second,_:=strconv.Atoi(duration[6:])
	//fmt.Printf("h=%d,m=%d,s=%d\n",hour,minute,second)
	totalSecond=hour*3600+minute*60+second
	return
}

//秒转为格式化数据
func translate2format(totalSecond int)(duration string){
	hour:=totalSecond/3600
	minute:=totalSecond%3600/60
	second:=totalSecond%60
	return fmt.Sprintf("%0d:%0d:%0d",hour,minute,second)
}

//使用ffmpeg读取文件，比较慢
func getDurationByFfmpeg(filePath string) int{
	cmd := exec.Command("ffmpeg", "-i",filePath)
	output, _ := cmd.CombinedOutput()
	outputString:=string(output)
	outputList:=strings.Split(outputString,"\r\n")
	//duration在第24个
	durationLine:=outputList[23]
	//解析提取数据:  Duration: 00:01:14.01, start: 0.080000, bitrate: 350 kb/s
	startP:=strings.Index(durationLine,":")
	endP:=strings.Index(durationLine,".")
	return translate2Second(durationLine[startP+2:endP])
}

//不使用ffmpeg，直接解析flv二进制文件，很快
func getDurationByReadFile(filePath string) int{
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	keyName:=""
	//一个字节一个字节读
	var currentByte = make([]byte, 0x01)
	var offset int64 = 0
	for {
		_, err = file.ReadAt(currentByte, offset)
		if err != nil {
			fmt.Println("读取文件时，发生错误！")
			break
		}
		//遇到小写字母，添加到keyName中
		if currentByte[0]>0x60 && currentByte[0]<0x7B{
			keyName+=string(currentByte[0])
		}else{
			//获取完整的keyName后判断是否拿到了duration
			if keyName=="duration"{
				//fmt.Println("找到了！")
				offset++
				break
			}else{
				keyName=""
			}
		}
		offset++
	}
	//拿到duration后，将数据按float64的格式读出
	var durationBytes=make([]byte,0x08)
	_, err = file.ReadAt(durationBytes, offset)
	var duration=math.Float64frombits(binary.BigEndian.Uint64(durationBytes))
	//方便计算，转成int
	return int(duration)
}



func main() {
	if debugMod{
		basePath="E:\\数据库"
		readMod=0
	}else{
		//"calculateProgress E:\\数据库 1"
		basePath=os.Args[1]
		if len(os.Args)>=3 && os.Args[2]=="1"{
			readMod=1
		}
	}
	var data=make(map[string]int)
	//分别获取看完的视频和没看完的视频的路径
	notYetList := getFilePath(basePath)
	passPath := fmt.Sprintf("%s\\pass",basePath)
	PassList:=getFilePath(passPath)

	infoPath:=fmt.Sprintf("%s\\%s",basePath,"info.txt")
	_,err:=os.Lstat(infoPath)
	if err!=nil{
		//首次运行，base目录下没有info文件的话，生成info文件
		fmt.Println("首次运行，获取数据")
		getInfo(data,notYetList)
		getInfo(data,PassList)
		saveData(data,infoPath)
	}else{
		fmt.Println("读取历史数据")
		data=loadData(infoPath)
	}
	//匹配文件名，并统计时长
	passTotal:=getTotalDuration(data,PassList)
	notYetTotal:=getTotalDuration(data,notYetList)
	fmt.Printf("已看过的视频总长度为：%s\n",translate2format(passTotal))
	fmt.Printf("还没看的视频总长度为：%s\n",translate2format(notYetTotal))
	a:=float64(passTotal)/float64(passTotal+notYetTotal)*100
	fmt.Printf("进度为%.2f%%\n",a)

}

func getTotalDuration(data map[string]int, filePathlist []string) int{
	totalDuration:=0
	for _,filePath:=range filePathlist{
		fileName:=getFileName(filePath)
		totalDuration+=data[fileName]
	}
	return totalDuration
}

