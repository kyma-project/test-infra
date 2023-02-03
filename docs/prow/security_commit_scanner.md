# Security Leaks Scanner

Welcome to the Security Leaks Scanner, a tool designed to scan repository for potential 
security leaks. This scanner operates using gitleaks, ensuring a thorough and efficient 
examination of your repository. With this tool, you can rest assured that your code base 
is protected against any potential security threats and vulnerabilities. Stay ahead of 
the game with the Security Leaks Scanner, your first line of defense against security breaches.

> **NOTE:** The Security Leaks Scanner is a mandatory evaluation that must be completed 
> prior to merging pull requests. This test is an essential part of ensuring the security 
> and integrity of the repository.

> **Important:** Gitleaks repository https://github.com/zricethezav/gitleaks

### How Security Leaks Scanner works

Every PR is examined for security leaks. Only the commits - changes to individual files are 
tested, not the entire repository or whole files. During the pull request testing process, 
gitleaks is executed, performing leak detection operations. The scanner takes into account 
commits from the 'main' branch to the last commit on your branch.

### Workflow

This is how the workflow looks like from the developer's perspective:

1. If you want, scan your commits locally, use 'gitleaks detect --log-opts="--all commitA..commitB"' 
Where commit A is SHA of main branch, and commit B is from top of your branch.
2. Submit a Pull Request.
3. Review the results.

### Failure in test

1. Identify the origin of the leak. The test result will indicate the location.
2. If you believe the leak was intentional and can be justified, add a comment to the 
line with leak '#gitleaks:allow'.
3. If the leak can be prevented, but has already been committed, use a squash commit or 
amend and push it to the remote branch.
4. If the leak persists even after removal, it remains in the branch history and the 
test will block the merge from completing.