# Alert

Will display a VSnackbar with the default settings Timeout `5000` and Btn `true`. TODO: config for the framework.

A complete axios response can be passed. if a json.error message exists, it will be displayed.

## Alert modes

Mode |
------ |
`ALERT`.`SUCCESS`      |
`ALERT`.`INFO`      |
`ALERT`.`ERROR`      |

## Trigger

```js
 store.commit('alert/' + ALERT.ERROR, "Message")
```
