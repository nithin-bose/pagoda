package handlers

import (
	"fmt"
	"time"

	"github.com/mikestefanello/backlite"
	"github.com/mikestefanello/pagoda/pkg/helpers"
	"github.com/mikestefanello/pagoda/pkg/msg"
	"github.com/mikestefanello/pagoda/templates/pages"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/mikestefanello/pagoda/pkg/form"
	"github.com/mikestefanello/pagoda/pkg/page"
	"github.com/mikestefanello/pagoda/pkg/services"
	"github.com/mikestefanello/pagoda/pkg/tasks"
)

const (
	routeNameTask       = "task"
	routeNameTaskSubmit = "task.submit"
)

type (
	Task struct {
		tasks *backlite.Client
		*services.TemplateRenderer
	}
)

func init() {
	Register(new(Task))
}

func (h *Task) Init(c *services.Container) error {
	h.TemplateRenderer = c.TemplateRenderer
	h.tasks = c.Tasks
	return nil
}

func (h *Task) Routes(g *echo.Group) {
	g.GET("/task", h.Page).Name = routeNameTask
	g.POST("/task", h.Submit).Name = routeNameTaskSubmit
}

func (h *Task) Page(ctx echo.Context) error {
	p := page.New(ctx)
	p.Title = "Create a task"

	p.TemplComponent = pages.Task(form.Get[helpers.TaskForm](ctx))

	return h.RenderPage(ctx, p)
}

func (h *Task) Submit(ctx echo.Context) error {
	var input helpers.TaskForm

	err := form.Submit(ctx, &input)

	switch err.(type) {
	case nil:
	case validator.ValidationErrors:
		return h.Page(ctx)
	default:
		return err
	}

	// Insert the task
	err = h.tasks.
		Add(tasks.ExampleTask{
			Message: input.Message,
		}).
		Wait(time.Duration(input.Delay) * time.Second).
		Save()

	if err != nil {
		return fail(err, "unable to create a task")
	}

	msg.Success(ctx, fmt.Sprintf("The task has been created. Check the logs in %d seconds.", input.Delay))
	form.Clear(ctx)

	return h.Page(ctx)
}
