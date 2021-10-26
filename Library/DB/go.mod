module MiaGame/Library/DB

go 1.15

require (
	MiaGame/Library/MiaError v0.0.0
	MiaGame/Library/MiaLog v0.0.0
	MiaGame/Library/MiaCrypt v0.0.0
	github.com/mediocregopher/radix/v3 v3.6.0
	github.com/og/x v0.0.0-20201210141255-dbe8c95570d3
	gorm.io/driver/mysql v1.0.3
	gorm.io/gorm v1.20.9
)

replace (
	MiaGame/Library/MiaError => ../MiaError
	MiaGame/Library/MiaLog => ../MiaLog
	MiaGame/Library/MiaCrypt => ../MiaCrypt

)
