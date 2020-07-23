module Test = {
  type http = {
    reqMethod: string,
    reqHeaders: option(Js.Dict.t(string)),
    reqBody: option(string),
    resStatus: option(int),
    resHeaders: option(Js.Dict.t(string)),
    resBody: option(string),
  };

  type promMetric = {
    key: string,
    lowerBound: float,
    upperBound: float,
    labels: list((Js.Dict.key, string)),
  };

  type portTest = {port: string};

  type promMetrics = {metrics: list(promMetric)};

  type dns = {
    record: string,
    strategy: string,
    check: list(string),
  };

  type ssh = {
    username: string,
    port: string,
    credentialType: string,
  };

  type specificTestInfo =
    | HTTP(http)
    | Prometheus(promMetrics)
    | TLS(portTest)
    | TCP(portTest)
    | DNS(dns)
    | SSH(ssh)
    | Empty;

  type t = {
    testId: string,
    testName: string,
    testType: string,
    timeout: int,
    url: string,
    interval: int,
    createdAt: string,
    specific: specificTestInfo,
  };
};

module Log = {
  type t = {
    logId: int,
    testId: string,
    statusId: int,
    message: string,
    responseTime: int,
    createdAt: string,
  };
};

module Contact = {
  type t = {
    contactId: string,
    contactName: string,
    contactType: string,
    contactUrl: string,
  };
};

module TestContact = {
  type t = {
    contactId: string,
    contactName: string,
    contactType: string,
    contactUrl: string,
    mutable threshold: int,
  };
};

module Decode = {
  let testSpecificInfo =
      (testType: string, json: Js.Json.t): Test.specificTestInfo => {
    Json.(
      switch (testType) {
      | "TLS" => Decode.(Test.TLS({port: json |> field("port", string)}))
      | "TCP" => Decode.(Test.TCP({port: json |> field("port", string)}))
      | "HTTP" =>
        Decode.(
          Test.HTTP({
            reqMethod: json |> field("req_method", string),
            reqHeaders:
              json |> field("req_headers", optional(dict(string))),
            reqBody: json |> field("req_body", optional(string)),
            resStatus: json |> field("res_status", optional(int)),
            resHeaders:
              json |> field("res_headers", optional(dict(string))),
            resBody: json |> field("res_body", optional(string)),
          })
        )
      | "Prometheus" =>
        let metricDecoder = json => {
          Test.{
            key: json |> Decode.field("key", Decode.string),
            lowerBound: json |> Decode.field("lower_bound", Decode.float),
            upperBound: json |> Decode.field("upper_bound", Decode.float),
            labels:
              (
                switch (
                  json
                  |> Decode.field(
                       "labels",
                       Decode.optional(Decode.dict(Decode.string)),
                     )
                ) {
                | Some(labels) => labels
                | None => Js.Dict.empty()
                }
              )
              |> Js.Dict.entries
              |> Array.to_list,
          };
        };
        Test.Prometheus({
          metrics:
            json |> Decode.field("metric_tests", Decode.list(metricDecoder)),
        });
      | "DNS" =>
        Test.DNS({
          record: json |> Decode.field("record", Decode.string),
          strategy: json |> Decode.field("strategy", Decode.string),
          check: json |> Decode.field("check", Decode.list(Decode.string)),
        })
      | "SSH" =>
        Test.SSH({
          username: json |> Decode.field("username", Decode.string),
          port: json |> Decode.field("port", Decode.string),
          credentialType:
            json |> Decode.field("credential_type", Decode.string),
        })

      | _ => Test.Empty
      }
    );
  };

  let test = json => {
    let testType = json |> Json.Decode.field("test_type", Json.Decode.string);
    Test.(
      Json.Decode.{
        testId: json |> field("test_id", string),
        testName: json |> field("test_name", string),
        testType,
        timeout: json |> field("timeout", int),
        url: json |> field("url", string),
        interval: json |> field("interval", int),
        createdAt: json |> field("created_at", string),
        specific: json |> field("blob", testSpecificInfo(testType)),
      }
    );
  };
  let tests = (json): array(Test.t) => Json.Decode.array(test, json);

  let log = json =>
    Log.(
      Json.Decode.{
        logId: json |> field("log_id", int),
        testId: json |> field("test_id", string),
        message: json |> field("message", string),
        statusId: json |> field("status_id", int),
        responseTime: json |> field("response_time", int),
        createdAt: json |> field("created_at", string),
      }
    );
  let logs = (json): list(Log.t) => Json.Decode.list(log, json);

  let contact = json => {
    Contact.(
      Json.Decode.{
        contactId: json |> field("contact_id", string),
        contactName: json |> field("contact_name", string),
        contactType: json |> field("contact_type", string),
        contactUrl: json |> field("contact_url", string),
      }
    );
  };
  let contacts = (json): list(Contact.t) => Json.Decode.list(contact, json);

  let testContact = json =>
    TestContact.(
      Json.Decode.{
        contactId: json |> field("contact_id", string),
        contactName: json |> field("contact_name", string),
        contactType: json |> field("contact_type", string),
        contactUrl: json |> field("contact_url", string),
        threshold: json |> field("threshold", int),
      }
    );
  let testContacts = (json): list(TestContact.t) =>
    Json.Decode.list(testContact, json);
};