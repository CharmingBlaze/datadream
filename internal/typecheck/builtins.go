package typecheck

// methodSpec describes a friendly namespace method.
type methodSpec struct {
	minArgs      int
	maxArgs      int // -1 = no fixed max
	optionFields map[string]bool
}

var shapeOpts = map[string]bool{
	"position": true, "center": true, "size": true,
	"width": true, "height": true, "radius": true, "color": true,
	"x": true, "y": true,
}

var drawTextOpts = map[string]bool{
	"position": true, "size": true, "color": true, "align": true,
	"font": true, "rotation": true, "shadow": true,
}

var drawTriangleOpts = map[string]bool{
	"p1": true, "p2": true, "p3": true, "a": true, "b": true, "c": true, "color": true,
}

var drawLineOpts = map[string]bool{
	"from": true, "to": true, "start": true, "end": true, "color": true,
}

var drawPointOpts = map[string]bool{
	"position": true, "point": true, "color": true,
}

var drawFpsOpts = map[string]bool{
	"position": true, "x": true, "y": true,
}

var uiWidgetOpts = map[string]bool{
	"position": true, "size": true, "width": true, "height": true, "x": true, "y": true,
}

var collisionRectOpts = map[string]bool{
	"position": true, "size": true, "width": true, "height": true,
}

var namespaces = map[string]map[string]methodSpec{
	"draw": {
		"text":          {minArgs: 2, maxArgs: 2, optionFields: drawTextOpts},
		"rect":          {minArgs: 1, maxArgs: 1, optionFields: shapeOpts},
		"sprite":        {minArgs: 1, maxArgs: 1},
		"fps":           {minArgs: 0, maxArgs: 1, optionFields: drawFpsOpts},
		"rectOutline":   {minArgs: 1, maxArgs: 1, optionFields: shapeOpts},
		"rectLines":     {minArgs: 1, maxArgs: 1, optionFields: shapeOpts},
		"circle":        {minArgs: 1, maxArgs: 1, optionFields: shapeOpts},
		"circleOutline": {minArgs: 1, maxArgs: 1, optionFields: shapeOpts},
		"circleLines":   {minArgs: 1, maxArgs: 1, optionFields: shapeOpts},
		"line":          {minArgs: 1, maxArgs: 1, optionFields: drawLineOpts},
		"triangle":      {minArgs: 1, maxArgs: 1, optionFields: drawTriangleOpts},
		"ellipse":       {minArgs: 1, maxArgs: 1, optionFields: shapeOpts},
		"point":         {minArgs: 1, maxArgs: 1, optionFields: drawPointOpts},
		"pixel":         {minArgs: 1, maxArgs: 1, optionFields: drawPointOpts},
	},
	"input": {
		"move2d":        {minArgs: 0, maxArgs: 0},
		"move3d":        {minArgs: 0, maxArgs: 0},
		"axis":          {minArgs: 4, maxArgs: 4},
		"pressed":       {minArgs: 1, maxArgs: 1},
		"down":          {minArgs: 1, maxArgs: 1},
		"released":      {minArgs: 1, maxArgs: 1},
		"mouse":         {minArgs: 0, maxArgs: 0},
		"mousePressed":  {minArgs: 0, maxArgs: 1},
		"mouseDown":     {minArgs: 0, maxArgs: 1},
		"mouseReleased": {minArgs: 0, maxArgs: 1},
		"scroll":        {minArgs: 0, maxArgs: 0},
		"wheel":         {minArgs: 0, maxArgs: 0},
	},
	"ui": {
		"button":      {minArgs: 1, maxArgs: 2, optionFields: uiWidgetOpts},
		"label":       {minArgs: 1, maxArgs: 2, optionFields: uiWidgetOpts},
		"labelButton": {minArgs: 1, maxArgs: 2, optionFields: uiWidgetOpts},
	},
	"audio": {
		"play":     {minArgs: 1, maxArgs: 1},
		"stop":     {minArgs: 1, maxArgs: 1},
		"unload":   {minArgs: 1, maxArgs: 1},
		"shutdown": {minArgs: 0, maxArgs: 0},
	},
	"assets": {
		"texture": {minArgs: 1, maxArgs: 1},
		"image":   {minArgs: 1, maxArgs: 1},
		"sound":   {minArgs: 1, maxArgs: 1},
		"unload":  {minArgs: 1, maxArgs: 1},
	},
	"collision": {
		"overlap":     {minArgs: 2, maxArgs: 2},
		"contains":    {minArgs: 2, maxArgs: 2},
		"pointInRect": {minArgs: 2, maxArgs: 2, optionFields: collisionRectOpts},
		"circle":      {minArgs: 3, maxArgs: 3},
	},
	"random": {
		"screenPosition": {minArgs: 0, maxArgs: 0},
		"point":          {minArgs: 0, maxArgs: 1},
		"int":            {minArgs: 0, maxArgs: 2},
		"float":          {minArgs: 0, maxArgs: 2},
	},
	"time": {
		"fps":     {minArgs: 0, maxArgs: 0},
		"now":     {minArgs: 0, maxArgs: 0},
		"elapsed": {minArgs: 0, maxArgs: 0},
		"frame":   {minArgs: 0, maxArgs: 0},
	},
	"math": {
		"dot":       {minArgs: 2, maxArgs: 2},
		"cross":     {minArgs: 2, maxArgs: 2},
		"normalize": {minArgs: 1, maxArgs: 1},
		"length":    {minArgs: 1, maxArgs: 1},
		"distance":  {minArgs: 2, maxArgs: 2},
		"lerp":      {minArgs: 3, maxArgs: 3},
		"clamp":     {minArgs: 3, maxArgs: 3},
	},
}

var namespaceRoots = map[string]bool{
	"draw": true, "input": true, "ui": true, "audio": true, "assets": true,
	"collision": true, "random": true, "time": true, "math": true,
	"screen": true, "keys": true, "colors": true,
}

var screenFields = map[string]bool{
	"width": true, "height": true, "center": true, "size": true,
}

var spriteFields = map[string]bool{
	"position": true, "scale": true, "rotation": true, "speed": true,
}

var builtinFns = map[string]bool{
	"vec2": true, "vec3": true, "vec4": true,
	"sprite": true, "Sprite": true, "sound": true,
	"clear": true, "quit": true, "print": true,
	"rgb": true, "rgba": true, "hsl": true, "css": true,
	"clamp": true, "lerp": true, "distance": true, "length": true, "normalize": true,
}
