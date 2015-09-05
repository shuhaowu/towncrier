package testhelpers

import "github.com/Sirupsen/logrus"

type LogrusTestHook struct {
	Logs map[logrus.Level][]*logrus.Entry
}

func NewLogrusTestHook() *LogrusTestHook {
	hook := &LogrusTestHook{}

	levels := hook.Levels()
	hook.Logs = make(map[logrus.Level][]*logrus.Entry, len(levels))
	for _, level := range levels {
		hook.Logs[level] = make([]*logrus.Entry, 0)
	}

	return hook
}

func (h *LogrusTestHook) Fire(entry *logrus.Entry) error {
	h.Logs[entry.Level] = append(h.Logs[entry.Level], entry)
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
	h.Logs = make(map[logrus.Level][]*logrus.Entry, len(h.Levels()))
}
