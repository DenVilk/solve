package invoker

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/udovin/algo/futures"
	"github.com/udovin/solve/managers"
	"github.com/udovin/solve/models"
	"github.com/udovin/solve/pkg"
	"github.com/udovin/solve/pkg/polygon"
)

type problemManager struct {
	cacheDir string
	files    *managers.FileManager
	problems map[int64]futures.Future[Problem]
	mutex    sync.Mutex
}

func newProblemManager(files *managers.FileManager, cacheDir string) (*problemManager, error) {
	if err := os.MkdirAll(cacheDir, os.ModePerm); err != nil {
		return nil, err
	}
	return &problemManager{
		cacheDir: cacheDir,
		files:    files,
		problems: map[int64]futures.Future[Problem]{},
	}, nil
}

func (m *problemManager) DownloadProblem(ctx context.Context, packageID int64) (Problem, error) {
	return m.downloadProblemAsync(ctx, packageID).Get(ctx)
}

func (m *problemManager) downloadProblemAsync(ctx context.Context, packageID int64) futures.Future[Problem] {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if problem, ok := m.problems[packageID]; ok {
		return problem
	}
	future, setResult := futures.New[Problem]()
	m.problems[packageID] = future
	go func() {
		problem, err := m.runDownloadProblem(ctx, packageID)
		if err != nil {
			m.deleteProblem(packageID)
		}
		setResult(problem, err)
	}()
	return future
}

func (m *problemManager) runDownloadProblem(ctx context.Context, packageID int64) (Problem, error) {
	problemFile, err := m.files.DownloadFile(ctx, packageID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = problemFile.Close() }()
	localProblemPath := filepath.Join(m.cacheDir, fmt.Sprintf("package-%d.zip", packageID))
	_ = os.Remove(localProblemPath)
	problemPath := filepath.Join(m.cacheDir, fmt.Sprintf("package-%d", packageID))
	_ = os.RemoveAll(problemPath)
	if file, ok := problemFile.(*os.File); ok {
		localProblemPath = file.Name()
	} else {
		localProblemFile, err := os.Create(localProblemPath)
		if err != nil {
			return nil, err
		}
		defer func() {
			_ = localProblemFile.Close()
			_ = os.Remove(localProblemPath)
		}()
		if _, err := io.Copy(localProblemFile, problemFile); err != nil {
			return nil, err
		}
		if err := localProblemFile.Close(); err != nil {
			return nil, err
		}
	}
	if err := pkg.ExtractZip(localProblemPath, problemPath); err != nil {
		return nil, fmt.Errorf("cannot extract problem: %w", err)
	}
	return &polygonProblem{path: problemPath}, nil
}

func (m *problemManager) deleteProblem(packageID int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	problemPath := filepath.Join(m.cacheDir, fmt.Sprintf("package-%d", packageID))
	_ = os.RemoveAll(problemPath)
	delete(m.problems, packageID)
}

type ProblemTest interface {
}

type ProblemTestGroup interface {
}

type ProblemStatement interface {
	Locale() string
	GetConfig() (models.ProblemStatementConfig, error)
}

type Problem interface {
	GetStatements() ([]ProblemStatement, error)
}

type polygonProblem struct {
	path   string
	config *polygon.Problem
}

func (p *polygonProblem) init() error {
	if p.config != nil {
		return nil
	}
	config, err := polygon.ReadProblem(p.path)
	if err != nil {
		return err
	}
	p.config = &config
	return nil
}

func (p *polygonProblem) GetStatements() ([]ProblemStatement, error) {
	if err := p.init(); err != nil {
		return nil, err
	}
	var statements []ProblemStatement
	for _, statement := range p.config.Statements {
		if statement.Type != "application/x-tex" {
			continue
		}
		if _, ok := polygonLocales[statement.Language]; !ok {
			continue
		}
		statements = append(statements, &polygonProblemStatement{
			problem:  p,
			language: statement.Language,
		})
	}
	return statements, nil
}

type polygonProblemStatement struct {
	problem  *polygonProblem
	language string
}

func (s *polygonProblemStatement) Locale() string {
	return polygonLocales[s.language]
}

func (s *polygonProblemStatement) GetConfig() (models.ProblemStatementConfig, error) {
	properties, err := polygon.ReadProblemProperites(
		s.problem.path, s.language,
	)
	if err != nil {
		return models.ProblemStatementConfig{}, err
	}
	config := models.ProblemStatementConfig{
		Locale: s.Locale(),
		Title:  properties.Name,
		Legend: properties.Legend,
		Input:  properties.Input,
		Output: properties.Output,
		Notes:  properties.Notes,
	}
	for _, sample := range properties.SampleTests {
		config.Samples = append(
			config.Samples,
			models.ProblemStatementSample{
				Input:  sample.Input,
				Output: sample.Output,
			},
		)
	}
	return config, nil
}

var polygonLocales = map[string]string{
	"russian": "ru",
	"english": "en",
}