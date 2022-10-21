# Tag Search App

## To run the application

* Copy contents of project folder – “src”, “files”, “templates” into go directory i.e. %USERPROFILE%\go in default installation of go
* Paste your API key in src\tagSearchApp\main.go at line 19
* run the following commands from your GOPATH set in the environment variables i.e. %USERPROFILE%\go in default installation of go
 * `go install tagSearchApp`
 * `.\bin\tagSearchApp.exe`
* Calls will be made to Clarifai's API to tag the images and once all images get tagged a success message will be displayed in the console and application can be accessed localhost:3000


