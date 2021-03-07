# Translate

The translations are checked in the folder `lang`. The files must be in the pattern `{lang}.json`. They are lazy loaded
by the application.

## Usage

Usage of the translation keys.

```js
$t('MessageID')
```

## api

Pattern | Method |Description |
------ |------ |------ |
| `/mode/overview` | `GET`  | will return the raw messages, available laguages, translated languages and the groups.|
| `/lang/{lang}/group/{group}` | `PUT`  | will create / update the group for this language. The json files will be updated.|
| `/lang/{lang}` | `DELETE`  | will delete the language. The json files will be deleted.|
| `/lang/{lang}.json` | `GET`  | return the json content.|

## used tranlsation keys:

Key |
------ |
|COMMON.Language |
|COMMON.Close |
|COMMON.Add |
|COMMON.Delete |
|COMMON.Save |
|COMMON.NoChanges |
|COMMON.DeleteItem |
|CONTROLLER.locale.Controller.Translation.AddLanguage |
|CONTROLLER.locale.Controller.Translation.Translation |
|CONTROLLER.locale.Controller.Translation.ID |
|CONTROLLER.locale.Controller.Translation.Title |
|CONTROLLER.locale.Controller.Translation.Description |

## I18nService

A service is defined which can be used by a vue app.

```js
import {I18nService} from 'gofer-vue'

let i18n = I18nService.i18n
new Vue({
    i18n,
    render: h => h(App)
}).$mount('#app')
```

The service has the option to lazy load languages. If a language was already loaded, no server request will be made.
Anyways, if you need to force a reload, set the second param to true.

```js
import {I18nService} from 'gofer-vue'

I18nService.loadLanguageAsync("en", false)
```

Can simple be used in `router.beforeEach`:

```js
router.beforeEach((to, from, next) => {
    const lang = to.params.lang
    I18nService.loadLanguageAsync(lang, false).then(() => next())
})
```
