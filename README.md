# ascii-draw

An ASCII art program written in Go, powered by the [tcell](https://github.com/gdamore/tcell) terminal package for rendering and input handling.

## Features

* Brush tool with a configurable brush radius size
* Line tool for making straight lines
* Undo-redo
* Lasso selection
* Drawings can be saved to a text file or to a custom format which preserves colors
* Copy, cut, and paste
* Can paste data directly from the user's clipboard into the program
* Color picking to grab colors and characters from the canvas
* Canvas resizing

## Upcoming Features

* Rectangle selection
* Exporting PNG images, letting user configure the font and color scheme
* A simple consumer library for saving, loading, and manipulating images in the ascii-draw format, and rendering with either Tcell or by outputting ANSI color codes
* An executable to print a file in the ascii-draw binary format to the terminal
* Additional selection transformations- scaling, shearing, and rotation.
* Instead of replacing selection, allow user to add, subtract, or intersect with the current selection
* Layer system
* Giving an actual name to this project

## Controls

q

## Limitations

Currently, to ensure maximum compatibility with all terminals, the program has the following limitations:

* The only characters permitted are the graphical 7-bit ASCII characters. These are the 95 characters from 0x20 (` `) to 0x7E (`~`) inclusive.
  ```
    ! " # $ % & ' ( ) * + , - . / 0 1 2
  3 4 5 6 7 8 9 : ; < = > ? @ A B C D E
  F G H I J K L M N O P Q R S T U V W X
  Y Z [ \ ] ^ _ ` a b c d e f g h i j k
  l m n o p q r s t u v w x y z { | } ~
  ```
* The only permissible foreground and background colors are the ANSI 4-bit colors (black, red, green, yellow, blue, magenta, cyan, white, plus their bright variants) and a blank color which represents the terminal's default.

Support for the full Unicode standard and full terminal colors may come in a future update.
