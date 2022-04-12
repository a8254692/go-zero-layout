package mylogrus

import (
	"context"
	"errors"
	"path"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

func InitLogger(pathStr string, pathNameStr string, level uint32, rotationTime uint32, keepDays uint32) (*logrus.Logger, error) {
	if pathStr == "" || pathNameStr == "" {
		return nil, errors.New("no path")
	}

	fileName := path.Join(pathStr, pathNameStr+".log")

	// 实例化
	logger := logrus.New()

	writer, err := rotatelogs.New(
		fileName+".%Y%m%d",
		// 生成软链，指向最新日志文件
		rotatelogs.WithLinkName(fileName),
		// 设置最大保存时间(天)
		rotatelogs.WithMaxAge(time.Duration(keepDays)*24*time.Hour),
		// 设置日志切割时间间隔(天)
		rotatelogs.WithRotationTime(time.Duration(rotationTime)*24*time.Hour),
	)
	if err != nil {
		return nil, err
	}

	logger.SetReportCaller(true)
	logger.SetOutput(writer)
	logger.SetLevel(logrus.Level(level))
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	//writeMap := lfshook.WriterMap{
	//	logrus.InfoLevel:  writer,
	//	logrus.FatalLevel: writer,
	//	logrus.DebugLevel: writer,
	//	logrus.WarnLevel:  writer,
	//	logrus.ErrorLevel: writer,
	//	logrus.PanicLevel: writer,
	//	logrus.TraceLevel: writer,
	//}
	//
	//lfHook := lfshook.NewHook(writeMap, &logrus.JSONFormatter{
	//	TimestampFormat: "2006-01-02 15:04:05",
	//})
	//
	//// 新增 Hook
	//logger.AddHook(lfHook)

	return logger, nil
}

//{
//"sender": "account-svc",           // 日志发送者
//"trace_id": "123456789",           // trace id
//"span_id": "123456789",            // span id
//"level": "ERROR",                  // 错误级别 ERROR/WARN/INFO/TRACK
//"code": "A0001",                   //  字母+4位数字, 字母代表错误源头
//"line": "xxxx.ts 10",              //  代码文件行
//"msg": "连接数据库失败",             // 错误描述
//"time": "2021-11-11 01:02:03.999",  // 时间
//"uin": "12345678",                // 迷你号（可选）
//"req": "xxx",                     // 请求消息内容  (可选)
//"resp": "xxx",                    // 响应消息内容 (可选)
//"track_data": {"event_name":"FinishOrder", "display_name":"完成订单","k1":"v1"} // level为TRACK必填,神策 (可选)
//}
func GetCommonField(ctx context.Context, fields map[string]interface{}) logrus.Fields {
	dataCopy := make(logrus.Fields)
	for k, v := range fields {
		dataCopy[k] = v
	}

	dataCopy["span_id"] = spanIdFromContext(ctx)
	dataCopy["trace_id"] = traceIdFromContext(ctx)

	return dataCopy
}

func GetCommonFieldNoTrace(fields map[string]interface{}) logrus.Fields {
	dataCopy := make(logrus.Fields)
	for k, v := range fields {
		dataCopy[k] = v
	}

	return dataCopy
}

func spanIdFromContext(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasSpanID() {
		return spanCtx.SpanID().String()
	}

	return ""
}

func traceIdFromContext(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		return spanCtx.TraceID().String()
	}

	return ""
}
