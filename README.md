# Mail Tester

Email testing for API based mail systems. Currently, only Mailgun is supported.

### Development

#### Live Reload

Install `gin`
```
go get github.com/codegangsta/gin
```

Run on port 4021, with a proxy from port 4020. This overrides the `.env` setting.
Visiting `localhost:4020` will trigger a rebuild if necessary, and automatically redirects to the app at `localhost:4021`
```
gin -a 4021 -p 4020 run mail.go
```