module MiaGame/Library/DB

go 1.15

require (
	MiaGame/Library/MiaCrypt v0.0.0
	MiaGame/Library/MiaError v0.0.0
	MiaGame/Library/MiaLog v0.0.0
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/lestrrat-go/file-rotatelogs v2.4.0+incompatible // indirect
	github.com/lestrrat-go/strftime v1.0.5 // indirect
	github.com/mediocregopher/radix/v3 v3.6.0
	github.com/og/x v0.0.0-20201210141255-dbe8c95570d3
	go.uber.org/zap v1.19.1 // indirect
	gorm.io/driver/mysql v1.0.3
	gorm.io/gorm v1.22.2
)

replace (
	MiaGame/Library/MiaCrypt => ../MiaCrypt
	MiaGame/Library/MiaError => ../MiaError
	MiaGame/Library/MiaLog => ../MiaLog

)
