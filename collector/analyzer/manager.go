package analyzer

import (
	"errors"
	"go.uber.org/zap"
)

type Manager map[Type]Analyzer

func NewManager(analyzers ...Analyzer) (Manager, error) {
	if len(analyzers) == 0 {
		return nil, errors.New("no analyzers found, but must provide at least one analyzer")
	}
	analyzerMap := make(map[Type]Analyzer)
	for i := range analyzers {
		analyzerMap[analyzers[i].Type()] = analyzers[i]
	}
	return analyzerMap, nil
}

func (manager Manager) GetAnalyzer(analyzerType Type) (Analyzer, bool) {
	analyzer, ok := manager[analyzerType]
	return analyzer, ok
}

func (manager Manager) StartAll(logger *zap.Logger) error {
	for _, analyzer := range manager {
		logger.Sugar().Infof("Starting analyzer [%s]", analyzer.Type())
		err := analyzer.Start()
		if err != nil {
			return err
		}
	}
	return nil
}

func (manager Manager) ShutdownAll(logger *zap.Logger) error {
	// There is possible that multiple errors will be returned, but now we only return one.
	// TODO: Combine multiple errors into one
	var err error = nil
	for _, analyzer := range manager {
		logger.Sugar().Infof("Shutdown analyzer [%s]", analyzer.Type())
		err = analyzer.Shutdown()
		if err != nil {
			logger.Sugar().Infof("Error shutting down analyzer: %v", err)
		}
	}
	return err
}
