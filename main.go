package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type status int

const divisor = 4

const (
    todo status = iota
    inProgress
    done
)

/* CUSTOM ITEM */
type Task struct {
    status status
    title string
    description string
}

func NewTask(status status, title, description string) Task {
    return Task{
        title: title, 
        description: description, 
        status: status,
    }
}

// TODO : Create a fix to move tasks to particular columns instead of the 
// following column by default

// func (t *Task) Prev() {
//     if t.status == todo {
//         t.status = done
//     } else {
//         t.status-- 
//     }
// }

func (t *Task) Next() {
    if t.status == done {
        t.status = todo
    } else {
        t.status++
    }
}

/* MODEL MANAGEMENT */
var models []tea.Model
const (
    model status = iota
    form
)

/* STYLING */
var (
    columnStyle = lipgloss.
        NewStyle().Padding(1, 2).
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("#3c3c3c"))
    focusedStyle = lipgloss.
        NewStyle().Padding(1, 2).
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("62"))
    helpStyle = lipgloss.
        NewStyle().Foreground(lipgloss.Color("241"))
)

// Implement the list.Item interface
func (t Task) FilterValue() string {
    return t.title
}

func (t Task) Title() string {
    return t.title
}

func (t Task) Description() string {
    return t.description
}

/* MAIN MODEL */
type Model struct {
    focused status
    lists   []list.Model
    err     error
    loaded  bool
    quit    bool
}

func New() *Model {
    return &Model{}
}

func (m *Model) MoveToNext() tea.Cmd {
    selectedItem := m.lists[m.focused].SelectedItem()
    selectedTask := selectedItem.(Task)
    m.lists[selectedTask.status].RemoveItem(m.lists[m.focused].Index())
    selectedTask.Next() // increment the status field
    m.lists[selectedTask.status].
        InsertItem(
            len(m.lists[selectedTask.status].Items()) - 1,
            list.Item(selectedTask),
        )
    return nil
}

func (m *Model) Next() {
    if m.focused == done {
        m.focused = todo
    } else {
        m.focused++
    }
}

func (m *Model) Prev() {
    if m.focused == todo {
        m.focused = done 
    } else {
        m.focused--
    }
}

func (m *Model) initLists(width, height int) {
    defaultList := list.New(
        []list.Item{}, 
        list.NewDefaultDelegate(), 
        width / divisor, 
        height - divisor,
    )
    defaultList.SetShowHelp(false)
    m.lists = []list.Model{defaultList, defaultList, defaultList}

    // Init To Do
    m.lists[todo].Title = "To Do"
    m.lists[todo].SetItems([]list.Item{
        Task {status: todo, title: "buy milk", description: "strawberry milk"},
        Task {status: todo, title: "eat sushi", description: "negitoro rolls, miso soup"},
        Task {status: todo, title: "fold laundry", description: "or wear wringly clothes"},
    });

    // Init In Progress
    m.lists[inProgress].Title = "In-Progress"
    m.lists[inProgress].SetItems([]list.Item{
        Task {status: inProgress, title: "stay cool", description: "as a cucumber"},
    });

    // Init Done
    m.lists[done].Title = "Done"
    m.lists[done].SetItems([]list.Item{
        Task {status: done, title: "write code", description: "Rust > Python"},
    });
}

func (m Model) Init() tea.Cmd {
    return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        if !m.loaded {
            columnStyle.Width(msg.Width/divisor)
            focusedStyle.Width(msg.Width/divisor)
            m.initLists(msg.Width, msg.Height)
            m.loaded = true
        }
        m.initLists(msg.Width, msg.Height)
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            m.quit = true
            return m, tea.Quit
        case "left", "h":
            m.Prev()
        case "right", "l":
            m.Next()
        case "enter":
            return m, m.MoveToNext() 
        case "n":
            models[model] = m // save the state of current model
            models[form] = NewForm(m.focused)
            return models[form].Update(nil) // renders the view as well
        }
    case Task:
        task := msg
        return m, m.lists[task.status].
            InsertItem(len(m.lists[task.status].Items()), task)
    }
    var cmd tea.Cmd
    m.lists[m.focused], cmd = m.lists[m.focused].Update(msg)
    return m, cmd
}

func (m Model) View() string {
    if !m.loaded {
        return "...loading"
    } 
    // Optimize for not having a final render
    if m.quit {
        return ""
    }
    todoView := m.lists[todo].View()
    inProgressView := m.lists[inProgress].View()
    doneView := m.lists[done].View()

    switch m.focused {
    case inProgress:
        return lipgloss.JoinHorizontal(
            lipgloss.Left, 
            columnStyle.Render(todoView),
            focusedStyle.Render(inProgressView),
            columnStyle.Render(doneView),
        )
    case done:
        return lipgloss.JoinHorizontal(
            lipgloss.Left, 
            columnStyle.Render(todoView),
            columnStyle.Render(inProgressView),
            focusedStyle.Render(doneView),
        )
    default:
        return lipgloss.JoinHorizontal(
            lipgloss.Left, 
            focusedStyle.Render(todoView),
            columnStyle.Render(inProgressView),
            columnStyle.Render(doneView),
        )
    }
}

/* FORM MODEL */
type Form struct {
    focused     status
    title       textinput.Model
    description textarea.Model
}

func NewForm(focused status) *Form {
    form := &Form{focused: focused}
    form.title = textinput.New()
    form.title.Focus()
    form.description = textarea.New()
    return form
}

func (m Form)CreateTask() tea.Msg {
    task := NewTask(m.focused, m.title.Value(), m.description.Value())
    return task
}

func (m Form) Init() tea.Cmd {
    return nil
}

func (m Form) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    switch msg := msg.(type) {

    // Navigating the model
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            return m, tea.Quit
        case "enter":
            // If you are done writing the title
            if m.title.Focused() {
                m.title.Blur()
                m.description.Focus()
                return m, textarea.Blink
            } else {
                models[form] = m
                return models[model], m.CreateTask
            }
        }
    }

    // Updating the title
    if m.title.Focused() {
        m.title, cmd = m.title.Update(msg)
        return m, cmd

    // Updating the description
    } else {
        m.description, cmd = m.description.Update(msg)
        return m, cmd
    }
}

func (m Form) View() string {
    return lipgloss.JoinVertical(lipgloss.Left, 
        m.title.View(), m.description.View())
}

func main() {
    models = []tea.Model{New(), NewForm(todo)}
    // Opens the main model by default
    m := models[model]    
    p := tea.NewProgram(m)
    if _, err := p.Run(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
