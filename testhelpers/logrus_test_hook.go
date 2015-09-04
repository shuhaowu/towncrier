package testhelpers

import "github.com/Sirupsen/logrus"

type LogrusTestHook struct {
	logs map[logrus.Level]*logrus.Entry
}

func NewLogrusTestHook() *LogrusTestHook {
	return &LogrusTestHook{
		logs: make(map[logrus.Level]*logrus.Entry),
	}
}

func (h *LogrusTestHook) Fire(entry *logrus.Entry) error {
	return nil
}

func (h *LogrusTestHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	}
}

func (h *LogrusTestHook) ClearLogs() {
	h.logs = make(map[logrus.Level]*logrus.Entry)
}
