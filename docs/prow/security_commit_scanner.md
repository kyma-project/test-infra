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

0. [optional] Scan your commits locally, use
'gitleaks detect --log-opts="--all commitA..commitB"' Where commit A is SHA of main branch, 
and commit B is from top of your branch.
1. Create a Pull Request.
2. Get the results. 
3. End - if the results are positive.

### Failure in test

1. Locate the source of the leak. The test result will indicate the location.
2. If you believe it is intentional and can be explained, add a comment to the 
commit with '#gitleaks:allow'. 
3. If you can avoid the leak, but it has already appeared in the commit history, 
you need to rewrite it. Use a squash commit or amend, and push it to your remote branch.
4. If not, even after removing the leak, it still exists in the history of your branch 
and the test will still not allow for the merge to be completed.