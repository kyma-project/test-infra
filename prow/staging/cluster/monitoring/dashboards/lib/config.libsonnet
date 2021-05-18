local util = import 'config_util.libsonnet';

//
// Edit configuration in this object.
//
local config = {
  local comps = util.consts.components,

  // Instance specifics
  instance: {
    name: "Kyma Prow",
    botName: "kyma-bot",
    url: "https://status-dev.prow.build.kyma-project.io",
    monitoringURL: "https://monitoring.status-dev.prow.build.kyma-project.io",
  },

  ciAbsents: {
    components: [
      comps.crier,
      comps.deck,
      comps.ghproxy,
      comps.hook,
      comps.horologium,
      comps.prowControllerManager,
      comps.sinker,
      comps.tide,
    ],
  },

  // How long we go during work hours without seeing a webhook before alerting.
  webhookMissingAlertInterval: '10m',

  // How many days prow hasn't been bumped.
  prowImageStaleByDays: {daysStale: 7, eventDuration: '24h'},
};

// Generate the real config by adding in constant fields and defaulting where needed.
{
  _config+:: util.defaultConfig(config),
  _util+:: util,
}
