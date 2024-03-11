# Documentation Guidelines

Follow the rules listed in this document to provide high-quality documentation.

- The repository must contain an automatically updated index page in the `docs` directory.
- The main `README.md` file of the repository must contain a link to the index page of the repository, where all documentation content is stored.
- The main `docs` directory must have a structure to group `.md` files according to their purpose. For example, there can be `what-is`, `how-to`, `architecture` directories and more.
- Each document file must have a prefix that reflects the purpose of the document, for example, `how-to_`, `what-is_`, or `architecture_`.
- Each `.md` file must begin with a title and a document description in the paragraph below. These two elements of an `.md` file are used to create the index page.
- All links in the first paragraph of an `.md` file must use absolute paths from the repository root directory.
- Documentation for specific tools and libraries must be maintained in the same directory as the tool or library.
- If a tool has a more complex documentation structure, it can have its own `docs` directory under its root directory.
- A `docs` directory under a repository root doesn't have to contain documentation related strictly to one tool only. For example, there can be documentation describing a system comprised of multiple tools which together deliver one service for users.
