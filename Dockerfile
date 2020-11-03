FROM pledgecamp/golang AS builder

ENV PROJECT_PATH $GOPATH/src/github.com/pledgecamp/pledgecamp-mail-tester
COPY . $PROJECT_PATH/

WORKDIR $PROJECT_PATH
RUN go install
EXPOSE 4020 

CMD ["/usr/local/go/bin/mail-tester"]
