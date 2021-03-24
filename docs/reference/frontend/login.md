# Login

The login has 3 forms.

* Login
* Forgot password
* Change password

Depending on the provider options `forgotPassword` or `changePassword`, these forms are available. If the option is not
given, the input fields and buttons are disabled.

An application can have one or more auth providers. If there are more than one configured, the user can select the
provider in the form. The provider information is added to all the requests for the backend.

At the moment the password field must have min 6 chars. TODO: create an option to set custom rules on each field.

## API calls

Func | Pattern|Description|
------ |------ |------ |
|login | /login (POST) | sends the param `login`,`password` and `provider` to the backend.|
|changePw | /pw/change (POST) | token,login,pw and provider as a param.|
|forgotPassword | /pw/forgot (POST) | login and provider as a param.|

## Translations

Used translation keys:

Keys |
------ |
`CONTROLLER.auth.Controller.Login.Description` |
`CONTROLLER.auth.Controller.Login.ForgotPassword`  |
`CONTROLLER.auth.Controller.Login.Privacy`  |
`CONTROLLER.auth.Controller.Login.Impress`  |
`CONTROLLER.auth.Controller.Login.ErrPasswordLength`     |
`CONTROLLER.auth.Controller.Login.ErrPasswordRequired`      |
`CONTROLLER.auth.Controller.Login.ErrPasswordMatch`      |
`CONTROLLER.auth.Controller.Login.ErrLoginRequired'`      |
`CONTROLLER.{{provider}}.ChangePassword.Success`      |
`CONTROLLER.{{provider}}.ChangePassword.Info`      |
`CONTROLLER.{{provider}}.ForgotPassword.Success`      |
`CONTROLLER.{{provider}}.ForgotPassword.Info`      |
`COMMON.Login`      |
`COMMON.Password'`   |
`COMMON.PasswordConfirm'`   |
`COMMON.Save'`   |
`COMMON.Reset`   |
`COMMON.Back`   |

