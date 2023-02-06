
var reporter = require('cucumber-html-reporter');

var options = {
        theme: 'bootstrap',
        jsonFile: 'test/report/cucumber_report.json',
        output: 'test/report/cucumber_report.html',
        reportSuiteAsScenarios: true,
        scenarioTimestamp: true,
        launchReport: true,
        brandTitle: "e2e test",
        storeScreenshots: false,
        scenarioTimestamp: true,
        metadata: {
            "Source Chain":"Gaia",
            "Target Chain":"Osmosis",
            "Test Environment": "Local Environment",
            "Relay": "Cosmos rly",
            "Platform": "linux",
            "Executed": "Local"
        }
    };

    reporter.generate(options);
    

    //more info on `metadata` is available in `options` section below.

    //to generate consodilated report from multi-cucumber JSON files, please use `jsonDir` option instead of `jsonFile`. More info is available in `options` section below.


    // Neha Kumar - nehakumari.sinha@icicibank.com  Ramya Shree - ramyashree.m@icicibank.com  // pon.karthika@icicibank.com