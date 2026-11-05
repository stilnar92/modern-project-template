# Learning while building

This folder is optional. Use it when a team member learns the stack through the project tasks. Agents write the core, and the learner reads what they did and why. The rules for agents are in the [AI standard](../dev/ai-standard/05-human-in-the-loop.md#learning-during-development). If nobody on the team learns this way, delete the folder and the debrief rule in `CLAUDE.md`.

| File | What it contains | Who keeps it |
| --- | --- | --- |
| [debriefs/](debriefs/README.md) | A debrief of each task: what we did, how, why, what is new, what appeared before, self-check questions | The agent, at the end of a task |
| [concepts.md](concepts.md) | The concept map: concepts and patterns, where they appeared, the level of understanding | The agent after a debrief. The learner confirms the level |
| [backend-path.md](backend-path.md) | The topic map by levels: order, sources, exercises | It grows with the stack |

## How one task teaches

1. **Before the task,** the agent reads the [concept map](concepts.md): what the learner has seen, understands, can do alone.
2. **During the task,** the agent can ask the learner to write a part. The part must be small and must have appeared at least once before. Hints come as a ladder: the idea, where to look, a piece of code.
3. **At the end,** the agent writes a [debrief](debriefs/README.md). First the big picture, then the details. Hard parts that the learner can skip for now are marked so.
4. **Repeats:** the agent links to the earlier debrief and says what is the same and what is different. Before it explains a repeated technique, it asks "how do you think we did it here?".
5. **The concept map** gets an update: the number of encounters and, when the learner explains or does the work alone, the level.

## Use past experience

- **Do not skip familiar topics, and do not explain them from zero. Check them.** Ask a short question first: "how did your previous stack do this?". The answer shows what stayed, what is forgotten and what is wrong.
- **Start with the common part:** what is the same as in a familiar stack, and what is different. Name the differences that are easy to miss.
- **A wrong idea:** the agent corrects it without judgment and records it in the [concept map](concepts.md#corrected-ideas): was → what is correct → where we found it.

## Adjusting difficulty

Difficulty changes in both directions.

| Signal | What the agent does next time |
| --- | --- |
| **Hard:** the third hint was necessary, a self-check answer was wrong, "I do not understand" | Adds the topic to [Review](concepts.md#review) with what was hard. Lowers the bar: a smaller part, a worked example first, hints at the start, a simpler explanation. The level can go one step down |
| **Easy:** done without hints, explained in own words | Raises the bar: a larger part, fewer hints, "why" questions instead of "what" questions |

- Come back to a hard topic soon: in the next task where it appears, or with a short exercise within 1–2 tasks.
- After a success at the lower bar, go up one step at a time.
- Difficulty is not a grade. It tells the agent how to explain next time.

## Levels of understanding

| Level | Meaning |
| --- | --- |
| Seen | It appeared in a debrief. The general idea is clear |
| Understand | The learner can explain what it is and why, in own words |
| Can do alone | The learner did it in a task or an exercise, without the third hint |

A pattern is fixed at "can do alone". A concept that appeared three times and stays at "seen" needs a separate exercise.

## Rules

- Learn on real code: exercises are parts of a task or small tasks next to the project code.
- There are no deadlines. Development sets the order. The [topic map](backend-path.md) only shows what comes first.
- An unclear explanation is a problem of the explanation. Ask for "simpler" or "with an example", and the agent explains again.
