# Security Leaks Scanner

Security Leaks Scanner is a tool that scans a repository for potential security leaks, thus providing protection against any potential security threats and vulnerabilities. It operates using [Gitleaks](https://github.com/zricethezav/gitleaks), which ensures a thorough and efficient examination of your repository. 

> **NOTE:** Running Security Leaks Scanner is mandatory and must be completed 
> before merging a pull request. It is essential for ensuring security 
> and integrity of a repository.

## How Security Leaks Scanner Works

Every PR is examined for security leaks. Only the commits - changes to individual files are tested, not the entire repository or whole files. During the pull request testing process, Gitleaks is executed, performing leak detection operations. The scanner takes into account commits from branch `main` to the last commit on your branch.

## Workflow

1. To scan your commits locally, use `gitleaks detect --log-opts="--all commitA..commitB"`,  where commit A is SHA of branch `main`, and commit B is from the top of your branch.
2. Submit a pull request.
3. Review the results.

## Failure in Test

Identify the origin of the leak. The test result will indicate the location. 

- If the leak is intentional and can be justified, add the `#gitleaks:allow` comment to the line with the leak.

- If the leak can be prevented but has already been committed, squash or amend the commit and push it to the remote branch.

- If the leak persists even after a removal attempt, it remains in the branch history and the test blocks the merge from completing.