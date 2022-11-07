# Documentation guidelines

1. Each repository should contain an automatically updated index page in `docs` directory.
2. The repository's main `README.md` file should contain a link to the repository's index page, where all documentation content is stored.
3. Each `.md` file should begin with a title and in the paragraph below, a document description . These two elements of `.md` file will be used to create index page.
4. A `docs` directory under the repository root doesn't have to contain documentation related strictly to the one tool only. For example, we can have a documentation describing a system comprised of multiple tools which together deliver some service for users.
5. Documentation for specific tools and libraries should be maintained in the same directory as the tool or library.
6. If a tool has a more complex documentation structure, it can have its own `docs` directory under its root directory.
7. Each document file should have a prefix that reflects the purpose of the document, for example: `how-to_`, `proposal_`, `standard_`, `architecture_`. If a document covers multiple topics, such as `how-to` and `architecture`, the prefix is not necessary.
8. The main `docs` directory should contain a structure to group `.md` files according to their purpose. For example, there could be `proposal`, `how-to`, `architecture` directories and more.
9. An automated index page from all repositories should be linked to the [Kyma Organisation](https://github.tools.sap/kyma/documentation) central documentation.
