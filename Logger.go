package masala

import ("flag"
	"fmt"
	"log"
	"os"
	"strconv")

type Log struct {
	Error error
	Message string
}

type Logger struct {
	writer	*log.Logger
}

func NewLogger() *Logger {
	flag.Parse()
	var file, error = os.OpenFile("../../log/src.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if error != nil {
		panic(error)
	}
	return &Logger{log.New(file, "", log.LstdFlags|log.Lshortfile)}
}

func (logger Logger) Error(error error) {
	if nil != error {
		fmt.Print(error.Error(), "\n")
		logger.writer.Output(2, error.Error())
	}
}

func (logger Logger) Float(value string) float64 {
	output, err := strconv.ParseFloat(value, 64)
	if nil != err {
		logger.writer.Output(2, "Conversion " + value + " to float failed")
		return 0
	}
	return output
}

func (logger Logger) Integer(value string) int {
	output, err := strconv.Atoi(value)
	if nil != err {
		logger.writer.Output(2, "Conversion of integer " + value + " failed")
		return 0
	}
	return output
}

func (logger Logger) Log(data Log) error {
	if nil != data.Error {
		logger.Message(data.Message)
	}
	return data.Error
}

func (logger Logger) Message(message string) Logger {
	logger.writer.Output(2, message)
	return logger
}

func (logger Logger) Panic(err error) {
	if nil != err {
		logger.writer.Output(2, err.Error())
		log.Panic(err)
	}
}