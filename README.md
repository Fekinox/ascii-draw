# ascii-draw

An ASCII art program written in Go, powered by the
[tcell](https://github.com/gdamore/tcell) terminal package for rendering
and input handling.

![completed calendar from advent of code
2016](https://github.com/user-attachments/assets/73e3e943-8f18-46f5-b70c-1fe19d9e285f)
*(Sample image is the completed calendar from [Advent of Code
2016](https://adventofcode.com/2016))*

## Features

- Brush tool with a configurable brush radius size
- Line tool for making straight lines
- Undo-redo
- Lasso selection
- Drawings can be saved to a text file or to a custom format which
  preserves colors
- Copy, cut, and paste
- Can paste data directly from the user's clipboard into the program
- Color picking to grab colors and characters from the canvas
- Canvas resizing

## Upcoming Features

- Rectangle selection
- Exporting PNG images, letting user configure the font and color scheme
- A simple consumer library for saving, loading, and manipulating images
  in the ascii-draw format, and rendering with either Tcell or by
  outputting ANSI color codes
- An executable to print a file in the ascii-draw binary format to the
  terminal
- Additional selection transformations- scaling, shearing, and rotation.
- Instead of replacing selection, allow user to add, subtract, or
  intersect with the current selection
- Layer system
- Giving an actual name to this project
- Persistent configuration

## Controls

| Key             | Command                                                                                                            |
|-----------------|--------------------------------------------------------------------------------------------------------------------|
| Any key         | Set the current brush character                                                                                    |
| Esc             | Return to brush tool                                                                                               |
| Ctrl+h          | Show help page                                                                                                     |
| Ctrl+q          | Quit                                                                                                               |
| Ctrl+Alt+q      | Quit without saving changes                                                                                        |
| Ctrl+f          | Select foreground color                                                                                            |
| Ctrl+g          | Select background color                                                                                            |
| Alt+=           | Increase brush radius                                                                                              |
| Alt+-           | Decrease brush radius                                                                                              |
| Alt+mouse hover | Look up character and colors on canvas                                                                             |
| Alt+click       | Grab character from canvas                                                                                         |
| Alt+drag up     | Grab foreground color                                                                                              |
| Alt+drag down   | Grab background color                                                                                              |
| Ctrl+z          | Undo                                                                                                               |
| Ctrl+y          | Redo                                                                                                               |
| Ctrl+c          | Copy selection                                                                                                     |
| Ctrl+x          | Cut selection                                                                                                      |
| Ctrl+v          | Paste selection                                                                                                    |
| Ctrl+Shift+v    | Paste clipboard                                                                                                    |
| Ctrl+a          | Reset selection                                                                                                    |
| Alt+,           | Clear selection                                                                                                    |
| Alt+.           | Fill selection                                                                                                     |
| Alt+1           | Toggle alpha lock: drawing commands do not modify empty characters (space ` ` characters with no background color) |
| Alt+2           | Toggle character lock: drawing commands do not change the character of a cell.                                     |
| Alt+3           | Toggle foreground lock: drawing commands do not change the foreground color of a cell.                             |
| Alt+4           | Toggle background lock: drawing commands do not change the background color of a cell.                             |

| Key    | Command               |
|--------|-----------------------|
| Ctrl+s | Save to binary file   |
| Ctrl+o | Load from binary file |
| Ctrl+i | Import plain text     |
| Ctrl+p | Export plain text     |

| Key                                       | Command                     |
|-------------------------------------------|-----------------------------|
| Esc                                       | Enter brush tool            |
| (Brush) Click and drag                    | Draw on canvas              |
| Ctrl+e                                    | Enter line tool             |
| (Line) Click and drag                     | Draw a straight line        |
| Ctrl+r                                    | Enter lasso tool            |
| (Lasso) Click and drag                    | Create a freeform selection |
| Ctrl+t                                    | Enter translate tool        |
| (Translate) Click and drag                | Move selected characters    |
| Alt+\[                                    | Enter resize tool           |
| (Resize) Click and drag outside of region | Set new resize rectangle    |
| (Resize) Click and drag inside of region  | Move resize bounds          |
| (Resize) Click and drag on edge of region | Move edge of resize region  |
| (Resize) Enter                            | Commit canvas resize        |

## Limitations

Currently, to ensure maximum compatibility with all terminals, the
program has the following limitations:

- The only characters permitted are the graphical 7-bit ASCII
  characters. These are the 95 characters from 0x20 (` `) to 0x7E (`~`)
  inclusive. ! " \# \$ % & ' ( ) \* + , - . / 0 1 2 3 4 5 6 7 8 9 : ; \<
  = \> ? @ A B C D E F G H I J K L M N O P Q R S T U V W X Y Z \[ \\ \]
  ^ \_ \` a b c d e f g h i j k l m n o p q r s t u v w x y z { \| } ~
- The only permissible foreground and background colors are the ANSI
  4-bit colors (black, red, green, yellow, blue, magenta, cyan, white,
  plus their bright variants) and a blank color which represents the
  terminal's default.

Support for the full Unicode standard and full terminal colors may come
in a future update.
