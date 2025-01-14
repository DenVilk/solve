package managers

import (
	"context"
	"database/sql"
	"sort"
	"time"

	"github.com/udovin/solve/core"
	"github.com/udovin/solve/models"
)

type ContestStandingsColumn struct {
	Problem models.ContestProblem
}

type ContestStandingsCell struct {
	Column  int
	Verdict models.Verdict
	Attempt int
	Time    int64
}

type ContestStandingsRow struct {
	Participant models.ContestParticipant
	Cells       []ContestStandingsCell
	Score       int
	Penalty     int64
}

type ContestStandings struct {
	Columns []ContestStandingsColumn
	Rows    []ContestStandingsRow
}

type ContestStandingsManager struct {
	contestParticipants *models.ContestParticipantStore
	contestSolutions    *models.ContestSolutionStore
	contestProblems     *models.ContestProblemStore
	solutions           *models.SolutionStore
}

func NewContestStandingsManager(core *core.Core) *ContestStandingsManager {
	return &ContestStandingsManager{
		contestParticipants: core.ContestParticipants,
		contestSolutions:    core.ContestSolutions,
		contestProblems:     core.ContestProblems,
		solutions:           core.Solutions,
	}
}

func (m *ContestStandingsManager) BuildStandings(ctx context.Context, contest models.Contest, now time.Time) (*ContestStandings, error) {
	contestConfig, err := contest.GetConfig()
	if err != nil {
		return nil, err
	}
	participants, err := m.contestParticipants.FindByContest(contest.ID)
	if err != nil {
		return nil, err
	}
	contestProblems, err := m.contestProblems.FindByContest(contest.ID)
	if err != nil {
		return nil, err
	}
	sortFunc(contestProblems, func(lhs, rhs models.ContestProblem) bool {
		return lhs.Code < rhs.Code
	})
	standings := ContestStandings{}
	columnByProblem := map[int64]int{}
	for i, problem := range contestProblems {
		standings.Columns = append(standings.Columns, ContestStandingsColumn{
			Problem: problem,
		})
		columnByProblem[problem.ID] = i
	}
	contestSolutions, err := m.contestSolutions.FindByContest(contest.ID)
	if err != nil {
		return nil, err
	}
	solutionsByParticipant := map[int64][]models.ContestSolution{}
	for _, solution := range contestSolutions {
		solutionsByParticipant[solution.ParticipantID] = append(
			solutionsByParticipant[solution.ParticipantID], solution,
		)
	}
	for _, participant := range participants {
		beginTime := int64(contestConfig.BeginTime)
		if participant.Kind == models.RegularParticipant {
			var participantConfig models.RegularParticipantConfig
			if err := participant.ScanConfig(&participantConfig); err != nil {
				continue
			}
			if participantConfig.BeginTime != 0 {
				beginTime = int64(participantConfig.BeginTime)
			}
		}
		participantSolutions, ok := solutionsByParticipant[participant.ID]
		if !ok {
			continue
		}
		solutionsByColumn := map[int][]models.Solution{}
		for _, participantSolution := range participantSolutions {
			solution, err := m.solutions.Get(participantSolution.SolutionID)
			if err != nil {
				if err == sql.ErrNoRows {
					continue
				}
				return nil, err
			}
			column, ok := columnByProblem[participantSolution.ProblemID]
			if !ok {
				continue
			}
			solutionsByColumn[column] = append(solutionsByColumn[column], solution)
		}
		row := ContestStandingsRow{
			Participant: participant,
		}
		for i := range standings.Columns {
			solutions, ok := solutionsByColumn[i]
			if !ok {
				continue
			}
			sortFunc(solutions, func(lhs, rhs models.Solution) bool {
				if lhs.CreateTime != rhs.CreateTime {
					return lhs.CreateTime < rhs.CreateTime
				}
				return lhs.ID < rhs.ID
			})
			cell := ContestStandingsCell{
				Column: i,
			}
			for _, solution := range solutions {
				if solution.CreateTime >= now.Unix() {
					continue
				}
				report, err := solution.GetReport()
				if err != nil {
					continue
				}
				if report == nil {
					cell.Attempt++
					cell.Verdict = 0
					break
				}
				if report.Verdict == models.CompilationError {
					continue
				}
				cell.Attempt++
				cell.Verdict = report.Verdict
				if beginTime != 0 {
					cell.Time = solution.CreateTime - beginTime
					if cell.Time < 0 {
						cell.Time = 0
					}
				}
				if report.Verdict == models.Accepted {
					break
				}
			}
			if cell.Attempt > 0 {
				row.Cells = append(row.Cells, cell)
			}
		}
		for _, cell := range row.Cells {
			if cell.Verdict == models.Accepted {
				problem := standings.Columns[cell.Column].Problem
				row.Score += getProblemScore(problem)
				row.Penalty += int64(cell.Attempt-1)*20 + cell.Time/60
			}
		}
		standings.Rows = append(standings.Rows, row)
	}
	sortFunc(standings.Rows, func(lhs, rhs ContestStandingsRow) bool {
		lhsOrder := getParticipantOrder(lhs.Participant.Kind)
		rhsOrder := getParticipantOrder(rhs.Participant.Kind)
		if lhsOrder != rhsOrder {
			return lhsOrder < rhsOrder
		}
		if lhs.Score != rhs.Score {
			return lhs.Score > rhs.Score
		}
		return lhs.Penalty < rhs.Penalty
	})
	return &standings, nil
}

func getParticipantOrder(kind models.ParticipantKind) int {
	switch kind {
	case models.ManagerParticipant:
		return 0
	case models.RegularParticipant:
		return 1
	default:
		return 2
	}
}

func getProblemScore(problem models.ContestProblem) int {
	config, err := problem.GetConfig()
	if err != nil {
		return 1
	}
	if config.Points != nil {
		return *config.Points
	}
	return 1
}

func sortFunc[T any](a []T, less func(T, T) bool) {
	impl := sortFuncImpl[T]{data: a, less: less}
	sort.Sort(&impl)
}

type sortFuncImpl[T any] struct {
	data []T
	less func(T, T) bool
}

func (s *sortFuncImpl[T]) Len() int {
	return len(s.data)
}

func (s *sortFuncImpl[T]) Swap(i, j int) {
	s.data[i], s.data[j] = s.data[j], s.data[i]
}

func (s *sortFuncImpl[T]) Less(i, j int) bool {
	return s.less(s.data[i], s.data[j])
}
