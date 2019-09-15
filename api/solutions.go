package api

import (
	"net/http"
	"sort"
	"strconv"

	"github.com/labstack/echo"

	"github.com/udovin/solve/models"
)

type Solution struct {
	models.Solution
	User    *models.User   `json:""`
	Problem *Problem       `json:""`
	Report  *models.Report `json:""`
}

func (v *View) GetSolution(c echo.Context) error {
	solutionID, err := strconv.ParseInt(c.Param("SolutionID"), 10, 64)
	if err != nil {
		c.Logger().Warn(err)
		return c.NoContent(http.StatusBadRequest)
	}
	solution, ok := v.buildSolution(solutionID)
	if !ok {
		return c.NoContent(http.StatusNotFound)
	}
	user, ok := c.Get(userKey).(models.User)
	if !ok {
		return c.NoContent(http.StatusForbidden)
	}
	if !v.canGetSolution(user, solution.Solution) {
		return c.NoContent(http.StatusForbidden)
	}
	return c.JSON(http.StatusOK, solution)
}

func (v *View) RejudgeSolution(c echo.Context) error {
	solutionID, err := strconv.ParseInt(c.Param("SolutionID"), 10, 64)
	if err != nil {
		c.Logger().Warn(err)
		return c.NoContent(http.StatusBadRequest)
	}
	solution, ok := v.buildSolution(solutionID)
	if !ok {
		return c.NoContent(http.StatusNotFound)
	}
	user, ok := c.Get(userKey).(models.User)
	if !ok {
		return c.NoContent(http.StatusForbidden)
	}
	if !user.IsSuper {
		return c.NoContent(http.StatusForbidden)
	}
	report := models.Report{
		SolutionID: solution.ID,
	}
	if err := v.app.Reports.Create(&report); err != nil {
		c.Logger().Error(err)
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, report)
}

func (v *View) GetSolutions(c echo.Context) error {
	user, ok := c.Get(userKey).(models.User)
	if !ok {
		return c.NoContent(http.StatusForbidden)
	}
	if !user.IsSuper {
		return c.NoContent(http.StatusForbidden)
	}
	var solutions []Solution
	for _, m := range v.app.Solutions.All() {
		if solution, ok := v.buildSolution(m.ID); ok {
			solution.SourceCode = ""
			if solution.Report != nil {
				solution.Report.Data = models.ReportData{}
			}
			solutions = append(solutions, solution)
		}
	}
	sort.Sort(solutionSorter(solutions))
	return c.JSON(http.StatusOK, solutions)
}

type reportDiff struct {
	Points  *float64 `json:""`
	Defense *int8    `json:""`
}

func (v *View) createSolutionReport(c echo.Context) error {
	solutionID, err := strconv.ParseInt(c.Param("SolutionID"), 10, 64)
	if err != nil {
		c.Logger().Warn(err)
		return c.NoContent(http.StatusBadRequest)
	}
	var diff reportDiff
	if err := c.Bind(&diff); err != nil {
		c.Logger().Warn(err)
		return c.NoContent(http.StatusBadRequest)
	}
	solution, ok := v.buildSolution(solutionID)
	if !ok {
		return c.NoContent(http.StatusNotFound)
	}
	user, ok := c.Get(userKey).(models.User)
	if !ok {
		return c.NoContent(http.StatusForbidden)
	}
	if !user.IsSuper {
		return c.NoContent(http.StatusForbidden)
	}
	report, ok := v.app.Reports.GetLatest(solution.ID)
	if !ok {
		return c.NoContent(http.StatusNotFound)
	}
	if diff.Defense != nil {
		report.Data.Defense = diff.Defense
	}
	if diff.Points != nil {
		report.Data.Points = diff.Points
	}
	if err := v.app.Reports.Create(&report); err != nil {
		c.Logger().Error(err)
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusCreated, report)
}

func (v *View) canGetSolution(
	user models.User, solution models.Solution,
) bool {
	if user.IsSuper {
		return true
	}
	if user.ID == solution.UserID {
		return true
	}
	if solution.ContestID > 0 {
		contest, ok := v.app.Contests.Get(solution.ContestID)
		if ok && user.ID == contest.UserID {
			return true
		}
	}
	return false
}

func (v *View) buildSolution(id int64) (Solution, bool) {
	solution, ok := v.app.Solutions.Get(id)
	if !ok {
		return Solution{}, false
	}
	result := Solution{
		Solution: solution,
	}
	if user, ok := v.app.Users.Get(solution.UserID); ok {
		result.User = &user
	}
	if problem, ok := v.buildProblem(solution.ProblemID); ok {
		problem.Description = ""
		result.Problem = &problem
	}
	if report, ok := v.app.Reports.GetLatest(solution.ID); ok {
		result.Report = &report
	}
	return result, true
}

type solutionSorter []Solution

func (c solutionSorter) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c solutionSorter) Len() int {
	return len(c)
}

func (c solutionSorter) Less(i, j int) bool {
	return c[i].ID > c[j].ID
}
