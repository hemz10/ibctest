Feature: Deploy SmartContract
    In order to connect 2 chains via IBC
    As a Relayer
    First step is to create clients on both chains

 Scenario: Deploying SmartContract on Osmosis
    Given Osmosis Chain running
    When we Deploy SmartContract on Osmosis
    Then Contract should be deployed on Osmosis