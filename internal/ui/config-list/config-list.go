package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dmitriy-rs/rollercoaster/internal/config"
	"github.com/dmitriy-rs/rollercoaster/internal/logger"
)

type model struct {
	keys         keyMap
	help         help.Model
	configs      []config.ConfigItem
	currentIndex int
	editing      bool
	selectIndex  int
	quitting     bool
	saved        bool
}

type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Enter  key.Binding
	Space  key.Binding
	Left   key.Binding
	Right  key.Binding
	Save   key.Binding
	Quit   key.Binding
	Toggle key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Space, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter, k.Space},
		{k.Left, k.Right, k.Quit},
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "toggle/edit"),
	),
	Space: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "toggle"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "previous option"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "next option"),
	),
	Save: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "save & exit"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q/esc", "save & exit"),
	),
}

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color("39"))

	editingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("39")).
			Bold(true)

	descriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241"))

	booleanStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86"))

	falseStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))
)

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case saveResult:
		m.saved = msg.success
		m.quitting = true
		return m, tea.Quit

	case tea.WindowSizeMsg:
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, m.saveConfig()

		case key.Matches(msg, m.keys.Save):
			return m, m.saveConfig()

		case key.Matches(msg, m.keys.Up):
			if !m.editing {
				if m.currentIndex > 0 {
					m.currentIndex--
				}
			}

		case key.Matches(msg, m.keys.Down):
			if !m.editing {
				if m.currentIndex < len(m.getVisibleConfigs())-1 {
					m.currentIndex++
				}
			}

		case key.Matches(msg, m.keys.Enter):
			if !m.editing {
				configItem := m.getCurrentConfig()
				if configItem.ItemType == "boolean" && configItem.Value != nil {
					// Toggle boolean directly
					visibleConfigs := m.getVisibleConfigs()
					configItem := visibleConfigs[m.currentIndex]
					configItem.Value = !configItem.Value.(bool)
					m.updateConfigInList(&configItem)
				} else if configItem.ItemType == "select" {
					// Open select mode
					m.editing = true
					// Find current selection index
					for i, option := range configItem.Options {
						if option == configItem.Value {
							m.selectIndex = i
							break
						}
					}
				}
			} else {
				m.editing = false
				// Apply the selected value
				visibleConfigs := m.getVisibleConfigs()
				configItem := visibleConfigs[m.currentIndex]
				if configItem.ItemType == "select" {
					configItem.Value = configItem.Options[m.selectIndex]
					m.updateConfigInList(&configItem)
				}
			}

		case key.Matches(msg, m.keys.Space):
			if !m.editing {
				configItem := m.getCurrentConfig()
				if configItem.ItemType == "boolean" && configItem.Value != nil {
					// Also allow space to toggle boolean for backwards compatibility
					visibleConfigs := m.getVisibleConfigs()
					configItem := visibleConfigs[m.currentIndex]
					configItem.Value = !configItem.Value.(bool)
					m.updateConfigInList(&configItem)
				}
			}

		case key.Matches(msg, m.keys.Left):
			if m.editing {
				configItem := m.getCurrentConfig()
				if configItem.ItemType == "select" && m.selectIndex > 0 {
					m.selectIndex--
				}
			}

		case key.Matches(msg, m.keys.Right):
			if m.editing {
				configItem := m.getCurrentConfig()
				if configItem.ItemType == "select" && m.selectIndex < len(configItem.Options)-1 {
					m.selectIndex++
				}
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.quitting {
		if m.saved {
			return "✓ Configuration saved successfully!\n"
		}
		return "Configuration not saved.\n"
	}

	s := titleStyle.Render("Rollercoaster Configuration") + "\n\n"

	visibleConfigs := m.getVisibleConfigs()
	for i, configItem := range visibleConfigs {
		cursor := "  "
		if i == m.currentIndex {
			cursor = "> "
		}

		var valueStr string
		var style lipgloss.Style

		if configItem.ItemType == "boolean" {
			if configItem.Value != nil && configItem.Value.(bool) {
				valueStr = booleanStyle.Render("✓ enabled")
			} else {
				valueStr = falseStyle.Render("✗ disabled")
			}
			style = itemStyle
		} else if configItem.ItemType == "select" {
			if m.editing && i == m.currentIndex {
				// Show all options with selection highlighting
				options := ""
				for j, option := range configItem.Options {
					if j == m.selectIndex {
						options += editingStyle.Render(fmt.Sprintf("[%s]", option))
					} else {
						options += fmt.Sprintf(" %s ", option)
					}
					if j < len(configItem.Options)-1 {
						options += " | "
					}
				}
				valueStr = options
				style = itemStyle // Use normal item style, not editing style for the whole line
			} else {
				if configItem.Value != nil {
					valueStr = booleanStyle.Render(configItem.Value.(string))
				} else {
					valueStr = booleanStyle.Render("unset")
				}
				style = itemStyle
			}
		}

		if i == m.currentIndex && !m.editing {
			style = selectedItemStyle
		}

		line := fmt.Sprintf("%s%s: %s", cursor, configItem.Label, valueStr)
		s += style.Render(line) + "\n"

		// Add description
		s += descriptionStyle.Render("    "+configItem.Description) + "\n\n"
	}

	s += "\n" + m.help.View(m.keys)

	if m.editing {
		s += "\n" + editingStyle.Render("EDITING MODE - Use ←/→ to change selection, Enter to confirm")
	} else {
		s += "\n" + descriptionStyle.Render("Changes are automatically saved when you exit")
	}

	return s
}

func (m model) getVisibleConfigs() []config.ConfigItem {
	var visible []config.ConfigItem
	for _, configItem := range m.configs {
		if configItem.Key == "DefaultJSManager" {
			// Only show if EnableDefaultJSManager is true
			enableJSManager := m.getConfigValue("EnableDefaultJSManager")
			if enableJSManager == nil || !enableJSManager.(bool) {
				continue
			}
		}
		visible = append(visible, configItem)
	}
	return visible
}

func (m model) getCurrentConfig() config.ConfigItem {
	visibleConfigs := m.getVisibleConfigs()
	return visibleConfigs[m.currentIndex]
}

func (m *model) updateConfigInList(updatedConfig *config.ConfigItem) {
	for i := range m.configs {
		if m.configs[i].Key == updatedConfig.Key {
			m.configs[i] = *updatedConfig
			break
		}
	}
}

func (m model) getConfigValue(key string) interface{} {
	for _, configItem := range m.configs {
		if configItem.Key == key {
			return configItem.Value
		}
	}
	return nil
}

func (m model) saveConfig() tea.Cmd {
	return func() tea.Msg {
		// Update all config values using the config package
		err := config.UpdateConfigItems(m.configs)
		if err != nil {
			logger.Error("Error saving config", err)
			return saveResult{success: false}
		}

		logger.Debug("Config saved successfully")
		return saveResult{success: true}
	}
}

type saveResult struct {
	success bool
}

func RenderConfigList() error {
	// Load current config values from config package
	configs := config.GetAllConfigItems()

	m := model{
		configs:      configs,
		currentIndex: 0,
		keys:         keys,
		help:         help.New(),
		editing:      false,
		selectIndex:  0,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	_, err := p.Run()
	if err != nil {
		return fmt.Errorf("error running config UI: %w", err)
	}

	return nil
}
