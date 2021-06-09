module github.com/patrickascher/gofer

go 1.16

replace github.com/spf13/viper => github.com/patrickascher/viper v1.7.1-mutex

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/fsnotify/fsnotify v1.4.9
	github.com/go-co-op/gocron v1.0.0
	github.com/go-gomail/gomail v0.0.0-20160411212932-81ebce5c23df
	github.com/go-playground/validator/v10 v10.6.1
	github.com/go-sql-driver/mysql v1.5.0
	github.com/guregu/null v4.0.0+incompatible
	github.com/imdario/mergo v0.3.12
	github.com/jinzhu/inflection v1.0.0
	github.com/julienschmidt/httprouter v1.3.1-0.20200921135023-fe77dd05ab5a
	github.com/mitchellh/mapstructure v1.3.3
	github.com/nicksnyder/go-i18n/v2 v2.1.2
	github.com/onsi/ginkgo v1.15.2 // indirect
	github.com/onsi/gomega v1.11.0 // indirect
	github.com/peterhellberg/duration v0.0.0-20191119133758-ec6baeebcd10
	github.com/rs/cors v1.7.0
	github.com/segmentio/ksuid v1.0.3
	github.com/serenize/snaker v0.0.0-20201027110005-a7ad2135616e
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/viper v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
	golang.org/x/text v0.3.5
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df // indirect
	gopkg.in/guregu/null.v4 v4.0.0
)
