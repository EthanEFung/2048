package main

import (
	"log"
	"math"
	"math/rand"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

/* node represents a grid coordinate pair within the "face" linked list */
type node struct {
	x, y int
	next *node
}

/*
coords is a node utility function that returns the x, y coordinate
tuple of the node.
*/
func (n node) coords() (x, y int) {
	return n.x, n.y
}

/*
face represents 4 linked lists of nodes that will help to manipulate
the grid.
*/
type face []*node

/*
board represents a grid of values the user manipulates, and has 4
faces representing a linked list "direction".
*/
type board struct {
	grid   [][]int
	top    face
	right  face
	bottom face
	left   face
}

/*
createBoard is a factory like function that returns a board grid
as well as attaching 4 faces to the board
*/
func createBoard() *board {
	result := new(board)

	grid := [][]int{}
	for i := 0; i < 4; i += 1 {
		row := make([]int, 4)
		grid = append(grid, row)
	}
	result.grid = grid

	top := []*node{}
	for i := 0; i < 4; i += 1 {
		list := &node{i, 0, nil}
		curr := list
		for j := 1; j < 4; j += 1 {
			curr.next = &node{i, j, nil}
			curr = curr.next
		}
		top = append(top, list)
	}
	result.top = top

	right := []*node{}
	for i := 0; i < 4; i += 1 {
		list := &node{3, i, nil}
		curr := list

		for j := 1; j < 4; j += 1 {
			curr.next = &node{3 - j, i, nil}
			curr = curr.next
		}
		right = append(right, list)
	}
	result.right = right

	bottom := []*node{}
	for i := 0; i < 4; i += 1 {
		list := &node{i, 3, nil}
		curr := list

		for j := 1; j < 4; j += 1 {
			curr.next = &node{i, 3 - j, nil}
			curr = curr.next
		}
		bottom = append(bottom, list)
	}
	result.bottom = bottom

	left := []*node{}
	for i := 0; i < 4; i += 1 {
		list := &node{0, i, nil}
		curr := list
		for j := 1; j < 4; j += 1 {
			curr.next = &node{j, i, nil}
			curr = curr.next
		}
		left = append(left, list)
	}
	result.left = left

	return result
}

/*
tilt is the board method that manipulates the values of the grid
given a face.
*/
func (b *board) tilt(face face) {
	grid := b.grid
	for _, list := range face {
		// per list, we will take all the values double the matched
		a, b := list, list
		for a != nil && b != nil {
			// first find the first non zero a
			for a != nil && grid[a.y][a.x] == 0 {
				a = a.next
			}
			if a == nil {
				// since there is no more, we allow the
				// loop to naturally exit
				continue
			}
			// now we find the first non zero neighbor
			b = a.next
			for b != nil && grid[b.y][b.x] == 0 {
				b = b.next
			}
			if b == nil {
				// since there is no more, we allow the
				// loop to naturally exit
				continue
			}
			// otherwise we should have a b with a non zero
			ax, ay := a.coords()
			bx, by := b.coords()
			if grid[ay][ax] == grid[by][bx] {
				grid[ay][ax] = 0
				grid[by][bx] *= 2
			}
			a = a.next
		}

		// per list we will then move shift all the values towards the
		// head of the face.
		a = list
		for i := 0; i < 3; i++ {
			ax, ay := a.coords()
			if grid[ay][ax] == 0 {
				b := a.next

				for b != nil {
					bx, by := b.coords()
					if grid[by][bx] != 0 {
						grid[ay][ax], grid[by][bx] = grid[by][bx], grid[ay][ax]
						break
					}
					b = b.next
				}
			}

			a = a.next
		}
	}
}

/*
place is a board method that randomly sets a "2" in an available
cell of the grid.
*/
func (b *board) place() {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)

	options := []node{}
	for y := range b.grid {
		for x := range b.grid[y] {
			if b.grid[y][x] == 0 {
				options = append(options, node{x, y, nil})
			}
		}
	}
	i := r.Intn(len(options))

	item := options[i]

	x, y := item.coords()

	b.grid[y][x] = 2
}

/*
changed is a board method that checks if any of the current grid values
differ from a previous board passed as an argument.
*/
func (b *board) changed(prev [][]int) bool {
	for y := range b.grid {
		for x := range b.grid[y] {
			if b.grid[y][x] != prev[y][x] {
				return true
			}
		}
	}
	return false
}

type model struct {
	board *board
}

/*
createModel is a factory like function that will create a board and place 2 2s
upon it, and return the model with the created board.
*/
func createModel() model {
	board := createBoard()
	board.place()
	board.place()
	return model{
		board: board,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// take a snapshot of the values before any effects take place
	prev := make([][]int, 4)
	for y := range m.board.grid {
		row := make([]int, 4)
		copy(row, m.board.grid[y])
		prev[y] = row
	}

	// read the event
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "left", "h":
			m.board.tilt(m.board.left)
		case "up", "k":
			m.board.tilt(m.board.top)
		case "right", "l":
			m.board.tilt(m.board.right)
		case "down", "j":
			m.board.tilt(m.board.bottom)
		}
	}

	// check to see if anything has changed on the board
	if m.board.changed(prev) {
		// the board should randomly generate a 2 on one of the open cells
		m.board.place()
	}
	return m, nil
}

var baseStyle = lipgloss.NewStyle().
	Bold(true).
	Underline(true)

/* cellStyle is a utility function that creates a lipgloss style */
func cellStyle(hex string) lipgloss.Style {
	return baseStyle.Copy().Foreground(lipgloss.Color(hex))
}

var styles = []lipgloss.Style{
	cellStyle("#eeeeee"), // 0
	cellStyle("#eee4da"), // 2
	cellStyle("#eee1c9"), // 4
	cellStyle("#f3b27a"), // 8
	cellStyle("#f69664"), // 16
	cellStyle("#f77c5f"), // 32
	cellStyle("#f75f3b"), // 64
	cellStyle("#edd073"), // 128
	cellStyle("#edcc62"), // 256
	cellStyle("#edc950"), // 512
	cellStyle("#edc53f"), // 1024
	cellStyle("#edc22e"), // 2048
}

func (m model) View() string {
	var s string
	grid := m.board.grid
	for y := range grid {
		for x := range grid[y] {
			f := float64(grid[y][x]) // used to determine which style of the styles variable should be used
			style := styles[0]
			if f > 0 {
				style = styles[int(math.Log2(f))]
			}

			num := strconv.Itoa(grid[y][x])
			s += style.Render(num)

			length := 6 - len(num)
			for i := 0; i < length; i += 1 {
				s += " "
			}
		}
		s += "\n\n"
	}
	s += "Use your arrow keys or h, j, k, l to move\nthe tiles. "
	s += "Tiles with the same number merge\ninto one when they touch. "
	s += "Add them up to\nreach 2048!\n"

	return s
}

func main() {
	program := tea.NewProgram(createModel())
	if err := program.Start(); err != nil {
		log.Fatalf("Bootup Error: %v", err.Error())
	}
}
