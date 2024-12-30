# labyrinth paper game

It's an implementation of [labyrinth](https://en.wikipedia.org/wiki/Labyrinth_(paper-and-pencil_game)) game.

# State of the development

Implemented:

 - earth cell
 - river cell
 - wormhole cell
 - solid walls
 - reading map as markdown table
 
TODO:

 - mechanics with treasure
 - shooting
 - hospital cell
 - arsenal cell
 - inner walls
 - bombs
 
 
# Use as helper tool for a master of the game

As a master you define a game with the maze map and list of players in a markdown file. I personally use VSCode with Markdown Table Formatter to make it easier. Here is a full example:

```
| X | 1 | 2     | 3     | 4     | 5  | 6     | 7 | 8 |
|---|---|-------|-------|-------|----|-------|---|---|
| 1 |   |       | W:A:1 |       |    |       |   |   |
| 2 |   |       |       |       |    |       |   | R |
| 3 |   |       |       | W:B:1 | RM | R     | R | R |
| 4 |   | W:A:0 |       |       |    |       |   |   |
| 5 |   |       |       |       |    |       |   |   |
| 6 |   |       |       |       |    |       |   |   |
| 7 |   |       | W:B:0 |       |    | W:A:2 |   |   |
| 8 |   |       |       |       |    |       |   |   |

exit: 9:8
alex: 2:2
tanya: 2:3

```

 - Use empty cell to define an earth.
 - `T:<system name>:<index>` means wormhole system. There are two wormhole systems (A, B) in the example above.
 - River is defined as `R` cells with `RM` as a river mouth. The tool will discover river flow by finding `RM`.
 - Solid walls will be generated automatically on each side of the maze.

Then you define an exit coordinates that must be placed on the solid wall `exit: row:column`. For example above other valid examples would be:

 - exit: 0:0
 - exit: 0:3
 - exit: 9:5

After defining the exit, you should write all players in format `<player name>: row:column`.