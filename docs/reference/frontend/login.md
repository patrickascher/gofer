# Login

## Env

Login is using the env `VUE_APP_BACKGROUND` and `VUE_APP_LOGO` to display the images. These env must be set or an error
will be shown.

## Translations

Used translation keys:

Keys |
------ |
`APPLICATION`.`Name`      |
`APPLICATION`.`Description`      |
`COMMON`.`Login`      |
`COMMON`.`Password`      |
`COMMON`.`ErrLoginRequired`      |
`COMMON`.`ErrPasswordRequired`      |
`COMMON`.`ErrPasswordLength`      |
`COMMON`.`Privacy`      |
`COMMON`.`Impress`      |

## User Service:

Func | Pattern|Description|
------ |------ |------ |
|login | /login (POST) | sends the param `login` and `password` to the backend.|
|logout | /logout (GET) | calls the url and deletes the user storage.`|

