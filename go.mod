module github.com/patrickascher/gofer

go 1.16

replace github.com/spf13/viper => github.com/patrickascher/viper v1.7.1-mutex

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/fsnotify/fsnotify v1.4.9
	github.com/go-playground/validator/v10 v10.4.1
	github.com/go-sql-driver/mysql v1.5.0
	github.com/guregu/null v4.0.0+incompatible
	github.com/jinzhu/inflection v1.0.0
	github.com/julienschmidt/httprouter v1.3.1-0.20200921135023-fe77dd05ab5a
	github.com/nicksnyder/go-i18n/v2 v2.1.2 // indirect
	github.com/patrickascher/gofw v0.1.10
	github.com/peterhellberg/duration v0.0.0-20191119133758-ec6baeebcd10
	github.com/rs/cors v1.7.0
	github.com/segmentio/ksuid v1.0.3
	github.com/serenize/snaker v0.0.0-20201027110005-a7ad2135616e
	github.com/sirupsen/logrus v1.8.0
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
	gopkg.in/guregu/null.v4 v4.0.0
)
