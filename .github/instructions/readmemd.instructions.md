---
description: Ensures consistency between README.md, copilot instructions, and prompt files.
applyTo: '**/*README.md'
---

Any significant changes done to README.md can mean we should also update copilot instructions and 
prompt files inside `./.github/` folder. 

That could also be true vice versa, changes to instructions and prompts could mean we need to update README.md but forgot.

Ensure that it is all in sync.
Check that paths, commands and other information are consistent in all these files.
