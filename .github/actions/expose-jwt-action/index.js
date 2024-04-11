const core = require('@actions/core');

async function run() {
  try {
    // Get aud and request token
    const audience = core.getInput('audience');
    console.log(`audience in the Javascript action: ${audience}`);
    const jwt = await core.getIDToken(audience);
    core.setOutput("jwt", "jwt");
  } catch (error) {
    core.setFailed(error.message);
  }
}

run()
