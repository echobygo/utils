package MiaLog

import (
    "io"
    "os"
    "strings"
    "time"

    rotatelogs "github.com/lestrrat-go/file-rotatelogs"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

var textLogger *zap.SugaredLogger
var jsonLogger *zap.SugaredLogger

func zapLogLevel(level string) zap.AtomicLevel {

    switch level {
    case "debug":
        return zap.NewAtomicLevelAt(zap.DebugLevel)
    case "info":
        return zap.NewAtomicLevelAt(zap.InfoLevel)
    case "warn":
        return zap.NewAtomicLevelAt(zap.WarnLevel)
    case "error":
        return zap.NewAtomicLevelAt(zap.ErrorLevel)
    }

    // default is info level
    return zap.NewAtomicLevelAt(zap.InfoLevel)
}
func InitLevelWithDay(level string,hourCount int){
    cfg := zapcore.EncoderConfig{
        MessageKey:  "msg",
        LevelKey:    "level",
        EncodeLevel: zapcore.CapitalColorLevelEncoder,
        TimeKey:     "timestamp",
        EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
            enc.AppendString(t.Format("2006-01-02 15:04:05"))
        },
        CallerKey:    "file",
        EncodeCaller: zapcore.ShortCallerEncoder,
        EncodeDuration: func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
            enc.AppendInt64(int64(d) / 1000000)
        },
    }

    // 设置一些基本日志格式 具体含义还比较好理解 直接看zap源码也不难懂
    textEncoder :=getEncoder();
    //

    // 实现两个判断日志等级的 interface
    //debugLevle := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
    //    return lvl == zapcore.DebugLevel
    //})
    infoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
        return lvl == zapcore.InfoLevel
    })
    infoLevel=infoLevel
    //warnLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
    //    return lvl == zapcore.WarnLevel
    //})
    errorLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
        return lvl == zapcore.ErrorLevel
    })
    // 打印所有级别的日志
    lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
        return lvl >= zapLogLevel(level).Level()
    })
    consoleDebugging := zapcore.Lock(os.Stdout)
    //jsonDebugWriter := getWriter("./logs/trade-debug.log")
    jsonInfoWriter := getWriterByTime("./logs/trade-info.log",hourCount)
    //jsonWarnWriter := getWriter("./logs/trade-warn.log")
    jsonErrorWriter := getWriterByTime("./logs/trade-error.log",hourCount)
    jsonEncoder := zapcore.NewJSONEncoder(cfg)
    jsonCore := zapcore.NewTee(
        zapcore.NewCore(jsonEncoder, zapcore.AddSync(jsonInfoWriter), infoLevel),
        zapcore.NewCore(jsonEncoder,consoleDebugging,lowPriority),
        zapcore.NewCore(jsonEncoder, zapcore.AddSync(jsonErrorWriter), errorLevel),
    )

    jsonLog := zap.New(jsonCore, zap.AddCaller(),zap.AddCallerSkip(2)) // 需要传入 zap.AddCaller() 才会显示打日志点的文件名和行数, 有点小坑
    jsonLogger = jsonLog.Sugar()

    ////获取 info、error日志文件的io.Writer 抽象 getWriter() 在下方实现
    //textDebugWriter := getWriter("./logs/trade-text-debug.log")
    textInfoWriter := getWriter("./logs/trade-text-info.log")
    //textWarnWriter := getWriter("./logs/trade-text-warn.log")
    textErrorWriter := getWriter("./logs/trade-text-error.log")

    //最后创建具体的 Logger
    textCore := zapcore.NewTee(
        zapcore.NewCore(textEncoder, zapcore.AddSync(textInfoWriter), infoLevel),
        zapcore.NewCore(textEncoder,consoleDebugging,lowPriority),
        zapcore.NewCore(textEncoder, zapcore.AddSync(textErrorWriter), errorLevel),
    )
    //

    testLog := zap.New(textCore, zap.AddCaller(),zap.AddCallerSkip(1)) // 需要传入 zap.AddCaller() 才会显示打日志点的文件名和行数, 有点小坑
    textLogger = testLog.Sugar()

}
func InitLevel(level string ){
    cfg := zapcore.EncoderConfig{
       MessageKey:  "msg",
       LevelKey:    "level",
       EncodeLevel: zapcore.CapitalColorLevelEncoder,
       TimeKey:     "timestamp",
       EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
           enc.AppendString(t.Format("2006-01-02 15:04:05"))
       },
       CallerKey:    "file",
       EncodeCaller: zapcore.ShortCallerEncoder,
       EncodeDuration: func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
           enc.AppendInt64(int64(d) / 1000000)
       },
    }

    // 设置一些基本日志格式 具体含义还比较好理解 直接看zap源码也不难懂
    textEncoder :=getEncoder();
   //

    // 实现两个判断日志等级的 interface
    //debugLevle := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
    //    return lvl == zapcore.DebugLevel
    //})
    infoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
        return lvl == zapcore.InfoLevel
    })
    infoLevel=infoLevel
    //warnLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
    //    return lvl == zapcore.WarnLevel
    //})
    errorLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
        return lvl == zapcore.ErrorLevel
    })
    // 打印所有级别的日志
    lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
        return lvl >= zapLogLevel(level).Level()
    })
    consoleDebugging := zapcore.Lock(os.Stdout)
    //jsonDebugWriter := getWriter("./logs/trade-debug.log")
    jsonInfoWriter := getWriter("./logs/trade-info.log")
    //jsonWarnWriter := getWriter("./logs/trade-warn.log")
    jsonErrorWriter := getWriter("./logs/trade-error.log")
    jsonEncoder := zapcore.NewJSONEncoder(cfg)
    jsonCore := zapcore.NewTee(
       zapcore.NewCore(jsonEncoder, zapcore.AddSync(jsonInfoWriter), infoLevel),
       zapcore.NewCore(jsonEncoder,consoleDebugging,lowPriority),
       zapcore.NewCore(jsonEncoder, zapcore.AddSync(jsonErrorWriter), errorLevel),
    )

    jsonLog := zap.New(jsonCore, zap.AddCaller(),zap.AddCallerSkip(2)) // 需要传入 zap.AddCaller() 才会显示打日志点的文件名和行数, 有点小坑
    jsonLogger = jsonLog.Sugar()

    ////获取 info、error日志文件的io.Writer 抽象 getWriter() 在下方实现
    //textDebugWriter := getWriter("./logs/trade-text-debug.log")
    textInfoWriter := getWriter("./logs/trade-text-info.log")
    //textWarnWriter := getWriter("./logs/trade-text-warn.log")
    textErrorWriter := getWriter("./logs/trade-text-error.log")

    //最后创建具体的 Logger
    textCore := zapcore.NewTee(
    	zapcore.NewCore(textEncoder, zapcore.AddSync(textInfoWriter), infoLevel),
        zapcore.NewCore(textEncoder,consoleDebugging,lowPriority),
    	zapcore.NewCore(textEncoder, zapcore.AddSync(textErrorWriter), errorLevel),
    )
    //

        testLog := zap.New(textCore, zap.AddCaller(),zap.AddCallerSkip(1)) // 需要传入 zap.AddCaller() 才会显示打日志点的文件名和行数, 有点小坑
        textLogger = testLog.Sugar()




}
func ShowBeegoTime(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
   // encodeTimeLayout(t, "2006-01-02T15:04:05.000Z0700", enc)
    enc.AppendString(t.Format("2006/01/02 15:04:05.000"))
}
func getEncoder() zapcore.Encoder {
    encoderConfig := zap.NewProductionEncoderConfig()
    encoderConfig.EncodeTime = ShowBeegoTime
    encoderConfig.EncodeLevel =zapcore.CapitalColorLevelEncoder
    encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
    return zapcore.NewConsoleEncoder(encoderConfig)


    //
    //return zapcore.EncoderConfig{
    //    // Keys can be anything except the empty string.
    //    TimeKey:        "T",
    //    LevelKey:       "L",
    //    NameKey:        "N",
    //    CallerKey:      "C",
    //    MessageKey:     "M",
    //    StacktraceKey:  "S",
    //    LineEnding:     zapcore.DefaultLineEnding,
    //    EncodeLevel:    zapcore.CapitalLevelEncoder,
    //    EncodeTime:     zapcore.ISO8601TimeEncoder,
    //    EncodeDuration: zapcore.StringDurationEncoder,
    //    EncodeCaller:   zapcore.ShortCallerEncoder,
    //}


//
}
func getWriterByTime(filename string,hourCount int) io.Writer {
    // 生成rotatelogs的 Logger 实际生成的文件名 trade-text-debug-YY-mm-dd-HH-MM-SS.log
    // trade_debug 是指向最新日志的链接
    // 保存7天内的日志 每1小时(整点)分割一次日志
    hook, err := rotatelogs.New(
        // 没有使用go风格反人类的format格式
        // strings.Replace(filename, ".log", "", -1)+"-%Y-%m-%d-%H-%M-%S.log",
        strings.Replace(filename, ".log", "", -1)+"-%Y-%m-%d.log",
        rotatelogs.WithLinkName(filename),
        rotatelogs.WithMaxAge(time.Hour*time.Duration(hourCount)*7),
        rotatelogs.WithRotationTime(time.Hour*time.Duration(hourCount)),  //24小时分割一次
    )

    if err != nil {
    }
    return hook
}
func getWriter(filename string) io.Writer {
    // 生成rotatelogs的 Logger 实际生成的文件名 trade-text-debug-YY-mm-dd-HH-MM-SS.log
    // trade_debug 是指向最新日志的链接
    // 保存7天内的日志 每1小时(整点)分割一次日志
    hook, err := rotatelogs.New(
        // 没有使用go风格反人类的format格式
        // strings.Replace(filename, ".log", "", -1)+"-%Y-%m-%d-%H-%M-%S.log",
        strings.Replace(filename, ".log", "", -1)+"-%Y-%m-%d.log",
        rotatelogs.WithLinkName(filename),
        rotatelogs.WithMaxAge(time.Hour*24*7),
        rotatelogs.WithRotationTime(time.Hour*24),  //24小时分割一次
    )

    if err != nil {
    }
    return hook
}

// Debug ....
func Debug(args ...interface{}) {
    jsonLogger.Debug(args...)
}

// Debugf ...
func Debugf(template string, args ...interface{}) {
    jsonLogger.Debugf(template, args...)
}

// Info ...
func Info(args ...interface{}) {
    jsonLogger.Info(args...)
}

// Infof ...
func Infof(template string, args ...interface{}) {
    jsonLogger.Infof(template, args...)
}

// Warn ...
func Warn(args ...interface{}) {
    jsonLogger.Warn(args...)
}

// Warnf ...
func Warnf(template string, args ...interface{}) {
    jsonLogger.Warnf(template, args...)
}

// Error ...
func Error(args ...interface{}) {
    jsonLogger.Error(args...)
}

// Errorf ...
func Errorf(template string, args ...interface{}) {
    jsonLogger.Errorf(template, args...)
}

// DPanic ...
func DPanic(args ...interface{}) {
    jsonLogger.DPanic(args...)
}

// DPanicf ...
func DPanicf(template string, args ...interface{}) {
    jsonLogger.DPanicf(template, args...)
}

// Panic ...
func Panic(args ...interface{}) {
    jsonLogger.Panic(args...)
}

// Panicf ...
func Panicf(template string, args ...interface{}) {
    jsonLogger.Panicf(template, args...)
}

// Fatal ...
func Fatal(args ...interface{}) {
    jsonLogger.Fatal(args...)
}

// Fatalf ...
func Fatalf(template string, args ...interface{}) {
    jsonLogger.Fatalf(template, args...)
}

// CDebug ....
func CDebug(args ...interface{}) {
    textLogger.Debug(args...)
}

// CDebugf ...
func CDebugf(template string, args ...interface{}) {
    textLogger.Debugf(template, args...)
}

// CInfo ...
func CInfo(args ...interface{}) {
    textLogger.Info(args...)
}

// CInfof ...
func CInfof(template string, args ...interface{}) {
    textLogger.Infof(template, args...)
}

// CWarn ...
func CWarn(args ...interface{}) {
    textLogger.Warn(args...)
}

// CWarnf ...
func CWarnf(template string, args ...interface{}) {
    textLogger.Warnf(template, args...)
}

// CError ...
func CError(args ...interface{}) {
    textLogger.Error(args...)
}

// CErrorf ...
func CErrorf(template string, args ...interface{}) {
    textLogger.Errorf(template, args...)
}

// CDPanic ...
func CDPanic(args ...interface{}) {
    textLogger.DPanic(args...)
}

// CDPanicf ...
func CDPanicf(template string, args ...interface{}) {
    textLogger.DPanicf(template, args...)
}

// CPanic ...
func CPanic(args ...interface{}) {
    textLogger.Panic(args...)
}

// CPanicf ...
func CPanicf(template string, args ...interface{}) {
    textLogger.Panicf(template, args...)
}

// CFatal ...
func CFatal(args ...interface{}) {
    textLogger.Fatal(args...)
}

// CFatalf ...
func CFatalf(template string, args ...interface{}) {
    textLogger.Fatalf(template, args...)
}
func GetTextLogger() *zap.SugaredLogger {
    return textLogger;
}