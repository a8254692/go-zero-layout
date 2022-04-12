##安装goctl对应模板已经修改过，默认模板请勿执行命令

##1.安装
>  框架安装  
  https://go-zero.dev/cn/prepare.html

>  **git分支最佳实践**  
  https://www.cnblogs.com/cnblogsfans/p/5075073.html

>  **go-zero开发规范**  
  https://go-zero.dev/cn/dev-specification.html


##2.项目架构
>  **命名规范**
  缩进:
  a. 使用四个空格
  makefile:  
  a. 变量以驼峰命名  
  b. 操作名以-为分隔符  
  c. 输出的可执行文件名以 _ 为分隔符  
  api:  
  a. xxx.api 文件中service命名格式为：apixxx(全小写)  
  b. http的path以/api/v1为前缀，第三位为服务标识 path均以下划线为分割  
  rpc:  
  a. rpcxxx.proto 文件均以rpc开头  

>  **目录结构**  
  /api  
  协议定义文件夹。  
  /configs  
  配置文件模板或默认配置。  
  /crontab  
  定时任务应用程序代码。  
  /deployments  
  发布需要文件系统和容器编排部署配置和模板(docker-compose、kubernetes/helm、mesos、terraform、bosh)  
  /docs  
  设计和用户文档  
  /service  
  私有应用程序和库代码。  
  /scripts  
  执行各种构建、安装、分析等操作的脚本。  
  /utils  
  脚手架目录。

>  **错误码详解**  
  `错误码分为两类 a.http状态码  b.系统内状态码(common/commstatus/openapicode.go)`  
  `存在两类的原因:如果出错了，仍然返回200状态码，有可能导致前端的处理发生混乱，这种情况要一定要禁止。特别是通用的API，基本上都是先看状态码再决定下一步的处理，如果没有返回正确的状态码，就会导致前端无法执行适当的方法去处理，从而引发各种不必要的问题。而且这种做法没有尽可能地运用HTTP协议，也给前端编写错误处理增加了难度。`  

>  **系统内状态码的详细内容**  
  `错误码组成：错误类型+应用标识+错误编码`  
  `错误码位数：7位`  
  `错误码示例：1000000`  
  `使用规范：只增不改、避免混乱、先占先得、写好注释`  
  `a.错误类型(2位数字,10开始)`  
  `例：参数错误：10`  
  `业务错误：11`  
  `b.应用标识(2位数字)`  
  `例：user-api：01`  
  `user-rpc：02`  
  `……`  
  `c.错误编码(3位数字)`

##模板修改
>  **进到项目根目录操作**  

>  a.执行 goctl template init  
  返回: Templates are generated in /root/.goctl/1.2.4-cli, edit on your risk!  
  
>  b.进入提示对应目录下的api  
  并执行 cp handler.tpl handler.tpl.bak1231

>  c.修改handler.tpl文件 参见: backups\goctl-template