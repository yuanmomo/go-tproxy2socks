#!/bin/sh      //解释所用脚本语言
# 添加定时任务到 crontab -e  "*/2 * * * * sh /root/monitor.sh" 

#************************************************
#函数:CheckProcess
#功能:检查一个进程是否运行正常
#参数:$1--要检查的进程名称
#返回:如果运行正常返回0,否则返回1

#************************************************
CheckProcess()
{
    # 检查输入的参数是否有效
    if [ "$1" = "" ];
    then
        return 1
    fi
    #$PROCESS_NUM获取指定进程的数量
    #如果值为1返回0,表示正常
    #如果不为1则返回1,表示有错误,需要重新启动
    #如果正常状态是多进程运行,则按进程数目修改
    ## "$1"为checkprocess函数代入参数（即可执行程序，本文指go-tproxy2socks）//这里我用的是ps

    PROCESS_NUM=`ps | grep "$1" | grep -v grep | wc -l`
    if [ $PROCESS_NUM -eq 1 ]; then
        return 0
    else
        return 1
    fi
}

#检查是否存在进程,这里引号里的部分参照自己的程序更改
while [ 1 ];do
    CheckProcess "/usr/bin/go-tproxy2socks";
    #$? 是shell标准变量,是上一个函数执行完毕return值
    Check_Result=$?
    if [ $Check_Result -eq 1 ]; then
        #有错误则杀死所有进程,如果并将标准输出及标准错误重定向到/dev/null
        #因为如果程序没有运行,进程数为0,你是无法kill的
        #然后重新启动进程
        killall -9 /usr/bin/go-tproxy2socks > /dev/null 2>&1
        /usr/bin/go-tproxy2socks -socks "socks5://10.10.1.4:1086"
    fi
    sleep 10
done

