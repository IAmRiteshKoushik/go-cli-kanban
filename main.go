package main

import(
  "github.com/charmbracelet/bubbles/list"
)

type status int

// Indices to determine the lists
const (
	todo status = iota
	inProgress
	done
)

/* CUSTOM ITEM */
type Task struct {
	status      status
	title       string
	description string
}

// implement the list.Item interface
func (t Task) FilterValue() string { // Filter based on title
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
  list list.Model
  err error
}

func (m *Model) initList() {
  m.list = list.New([]list.Item{}, list.NewDefaultDelegate())
}
