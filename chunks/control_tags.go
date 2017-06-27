package chunks

import (
	"fmt"
	"io"
)

// controlTagDefinitions is a map of tag names to control tag definitions.
var controlTagDefinitions = map[string]*controlTagDefinition{}

// ControlTagAction runs the interpreter.
type ControlTagAction func(ASTControlTag) func(io.Writer, Context) error

// controlTagDefinition tells the parser how to parse control tags.
type controlTagDefinition struct {
	name        string
	isBranchTag, isEndTag bool
	syntaxModel *controlTagDefinition
	parent      *controlTagDefinition
	action      ControlTagAction
}

func (c *controlTagDefinition) compatibleParent(p *controlTagDefinition) bool {
	if p == nil {
		return false
	}
	if p.syntaxModel != nil {
		p = p.syntaxModel
	}
	return c.parent == p
}

func (c *controlTagDefinition) requiresParent() bool {
	return c.isBranchTag || c.isEndTag
}

func (c *controlTagDefinition) isStartTag() bool {
	return !c.isBranchTag && !c.isEndTag
}

func addControlTagDefinition(ct *controlTagDefinition) {
	controlTagDefinitions[ct.name] = ct
}

func findControlTagDefinition(name string) (*controlTagDefinition, bool) {
	ct, found := controlTagDefinitions[name]
	return ct, found
}

type tagBuilder struct {tag *controlTagDefinition}

// DefineControlTag defines a control tag and its matching end tag.
func DefineControlTag(name string) tagBuilder {
	ct := &controlTagDefinition{name: name}
	addControlTagDefinition(ct)
	addControlTagDefinition(&controlTagDefinition{name: "end" + name, isEndTag: true, parent: ct})
	return tagBuilder{ct}
}

// Branch tells the parser that the named tag can appear immediately between this tag and its end tag,
// so long as it is not nested within any other control tags.
func (b tagBuilder) Branch(name string) tagBuilder {
	addControlTagDefinition(&controlTagDefinition{name: name, isBranchTag: true, parent: b.tag})
	return b
}

// Governs tells the parser that the tags can appear anywhere between this tag and its end tag.
func (b tagBuilder) Governs(_ []string) tagBuilder {
	return b
}

// SameSyntaxAs tells the parser that this tag has the same syntax as the named tag.
func (b tagBuilder) SameSyntaxAs(name string) tagBuilder {
	ot := controlTagDefinitions[name]
	if ot == nil {
		panic(fmt.Errorf("undefined: %s", name))
	}
	b.tag.syntaxModel = ot
	return b
}

// Action sets the action for a control tag definition.
func (b tagBuilder) Action(fn ControlTagAction) {
	b.tag.action = fn
}