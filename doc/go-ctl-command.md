##1.创建api
>   **进到 ./api/back-common-protocol/rulesapi/ 目录执行**

>   生成网关Api的代码（goctl对应模板已经修改过，默认模板请勿执行下面命令）:  
    `goctl api go -api userbehavior.api -dir ../../../service/userbehavior/api/`

##2.创建rpc
>   **进到 ./api/back-common-protocol/proto/ 目录操作**

>   生成 userbehavior 的RPC的代码:  
    `goctl rpc proto -src rpcuserbehavior.proto -dir ../../../service/userbehavior/rpc/`   

    `goctl rpc protoc rpcuserbehavior.proto --go_out=../../../service/userbehavior/rpc/pb  --go-grpc_out=../../../service/userbehavior/rpc/pb  --zrpc_out=../../../service/userbehavior/rpc/`

##3.创建model
>   **进到 service/userbehavior/model/ 目录操作**

>   生成 mongo model 示例代码: (先创建type.go文件并填入对应结构体)   
    `goctl model mongo -t Gift -style gozero -dir ./giftmodel`

>   生成 mysql model 示例代码:  
    `goctl model mysql ddl -c -src book.sql -dir .`  
    `goctl model mysql datasource -url="root:123456@tcp(127.0.0.1:3306)/gozero" -table="sys*" -dir ./sysmodel`


>   生成 user_count model 代码:  
    `goctl model mysql datasource -url="root:123456@tcp(10.0.0.106:3306)/sirius" -table="user_count" -dir ./usercount`

>   生成 user_focus model 代码:  
    `goctl model mysql datasource -url="root:123456@tcp(10.0.0.106:3306)/sirius" -table="user_focus" -dir ./userfocus`

>   生成 user_attention model 代码:  
    `goctl model mysql datasource -url="root:123456@tcp(10.0.0.106:3306)/sirius" -table="user_praise" -dir ./userpraise`

>   生成 comment model 代码:  
    `goctl model mysql datasource -url="root:123456@tcp(10.0.0.106:3306)/sirius" -table="user_focus_msg_log" -dir ./userfocusmsglog`

>   生成 comment model 代码:  
    `goctl model mysql datasource -url="root:123456@tcp(10.0.0.106:3306)/sirius" -table="comment" -dir ./comment`

>   生成 usermgo model 代码:  
    `goctl model mongo -t User -style gozero -dir ./usermgo`