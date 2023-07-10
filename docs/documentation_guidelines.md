# Documentation guidelines

1. Each repository must contain an automatically updated index page in `docs` directory.
2. An automated index page from all repositories must be linked to the [Kyma Organisation](https://github.tools.sap/kyma/documentation) documentation.
3. The main `README.md` file of a repository must contain a link to the index page of the repository, where all documentation content is stored.
4. The main `docs` directory must have a structure to group `.md` files according to their purpose. For example, there can be `what-is`, `how-to`, `architecture` directories and more.
5. Each document file must have a prefix that reflects the purpose of the document, for example, `how-to_`, `what-is_`, or `architecture_`.
6. Each `.md` file must begin with a title and a document description in the paragraph below. These two elements of `.md` file are used to create the index page.
7. All links in the first paragraph of `.md` must use absolute paths from the repository root directory.
8. Documentation for specific tools and libraries must be maintained in the same directory as the tool or library.
9. If a tool has a more complex documentation structure, it can have its own `docs` directory under its root directory.
10. A `docs` directory under a repository root doesn't have to contain documentation related strictly to one tool only. For example, there can be documentation describing a system comprised of multiple tools which together deliver one service for users.
