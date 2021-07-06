# Go解析Flv数据，统计看视频进度

B站是个好地方，上面有好多视频可以学习，但有时候网络不稳定，于是便使用工具将视频下载了下来。在看视频学习的过程中，想统计一下自己的学习进度，便开发了这个工具

github地址：https://github.com/Andrew201801/calculateProgress

## 使用教程：

* 下载release文件，并解压

* 在放视频的文件夹下，新建一个pass目录，将看过的视频放进去

  * 如下例，数据库文件夹下，有《149什么是函数依赖.flv》等一堆看过的文件，而在数据库文件夹下的pass文件夹内，包含了《 001数据库系统课程简要介绍.flv》等一堆看过的文件
  * ![](https://gitee.com/AndrewYu/PicGo/raw/master/img/20210706155345.png)
  * ​		...

* cd 进release文件的目录，使用命令行调用：

  ```
  calculateProgress.exe 目录名称 读取模式
  
  如上例，调用命令为(三者都可)：
  calculateProgress E:\\数据库 0
  calculateProgress E:\\数据库 1
  calculateProgress E:\\数据库
  ```

  目录名称：视频文件的根目录

  读取模式：默认为0，可以为空，0表示直接读取文件，比较快，1表示使用ffmpeg调用，比较慢

  首次使用将生成info.txt文件，记录视频的时长数据。后续使用将直接使用这个文件，可以直接获得数据。

  

## 开发过程&代码注释

其实去年考研期间用python写过一个一样的工具，这次怀着练习go的语法的想法，用go重构了一下，并深入学习了mp4和flv的文件格式。

整体流程是这样的：

* 读取看完和没看完的视频的路径
* 检查是否首次运行，首次运行就读取解析flv文件，生成info.txt文件
* 将读取到的视频时长数据和看完，没看完的视频进行匹配，统计各自的时长
* 将统计结果输出

可以看出来，整体流程来说并没有什么难点，关键在于如何读取flv文件，获取视频的时长数据，我是这么做的（不想看可以直接跳过，看下面的成品）：

* 网上搜索“go 视频 时长”，拿到了这篇文章：[纯Golang获取MP4视频时长信息](https://www.cnblogs.com/Akkuman/p/12371838.html)，于是我为了弄懂代码里的含义，去看了MP4文件格式的解析，具体来说就是一堆堆的box,然后时长数据在moov box里，使用timeScale和Duration字段的值，就能计算出来，但我要白嫖代码的时候，发现我的视频都是flv格式的ORZ，于是这条路堵死了。不过这个代码也对我有很大帮助，我从中学习到了如何读取二进制文件
* 之后我开始查找flv文件的格式，找到了[flv格式详解+实例剖析](https://www.jianshu.com/p/7ffaec7b3be6)，在博主的帮助下，我知道了时长数据在首个tag，onMetaData里的duration字段里，并需要按照double(Golang是float64)的方式解析，再参考[Golang 中的常见字节操作](https://blog.csdn.net/oyoung_2012/article/details/107712240)，知道了如何将[]byte转为double

成品：

示例，如下面这个文件：

![image-20210706152731875](https://gitee.com/AndrewYu/PicGo/raw/master/img/image-20210706152731875.png)

时长数据为：40 81 BC 0A 3D 70 A3 D7(都是十六进制，需要解析为double)

代码：

```go
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
```

(PS:double的解析也可以自己来，但需要计组基础，[IEEE 754浮点数标准详解](http://c.biancheng.net/view/314.html))



引用：
[纯Golang获取MP4视频时长信息](https://www.cnblogs.com/Akkuman/p/12371838.html)
[flv格式详解+实例剖析](https://www.jianshu.com/p/7ffaec7b3be6)
[Golang 中的常见字节操作](https://blog.csdn.net/oyoung_2012/article/details/107712240)
[IEEE 754浮点数标准详解](http://c.biancheng.net/view/314.html)