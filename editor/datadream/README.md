# DataDream for VS Code

Syntax highlighting and a focused dark editor theme for `.dd` files.

## Theme

This extension includes **DataDream Studio**, a dark theme tuned for DataDream code:

- restrained dark workbench colors
- clearer active tabs, panels, sidebars, and status bar states
- softer line numbers and indentation guides
- brighter, more readable DataDream syntax colors

After installing the extension, open the command palette and choose:

```text
Preferences: Color Theme -> DataDream Studio
```

## Install (development)

1. Open this folder (`editor/datadream`) in VS Code.
2. Press F5 to launch an Extension Development Host.
3. Open any `.dd` file.
4. Select the **DataDream Studio** color theme.

## Install (from repo)

Package and install locally:

```bash
cd editor/datadream
npm install -g @vscode/vsce   # once
vsce package
code --install-extension datadream-0.1.0.vsix
```

## Standalone TextMate grammar

Editors that accept `.tmLanguage.json` can use:

`syntaxes/datadream.tmLanguage.json` (repo root)
