module Test = {
  type http = {
    method_: string,
    payload: option(string),
    expResult: option(string),
  };

  type promMetric = {
    key: string,
    lowerBound: float,
    upperBound: float,
    labels: option(Js.Dict.t(string)),
  };

  type portTest = {port: int};

  type promMetrics = {metrics: array(promMetric)};

  type specificTestInfo =
    | HTTP(http)
    | Prometheus(promMetrics)
    | TLS(portTest)
    | TCP(portTest)
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
      | "TLS" => Decode.(Test.TLS({port: json |> field("port", int)}))
      | "TCP" => Decode.(Test.TCP({port: json |> field("port", int)}))
      | "HTTP" =>
        Decode.(
          Test.HTTP({
            method_: json |> field("method", string),
            expResult: json |> field("exp_result", optional(string)),
            payload: json |> field("payload", optional(string)),
          })
        )
      | "Prometheus" =>
        let metricDecoder = json => {
          Test.{
            key: json |> Decode.field("key", Decode.string),
            lowerBound: json |> Decode.field("lower_bound", Decode.float),
            upperBound: json |> Decode.field("upper_bound", Decode.float),
            labels:
              json
              |> Decode.field(
                   "labels",
                   Decode.optional(Decode.dict(Decode.string)),
                 ),
          };
        };
        Test.Prometheus({
          metrics:
            json |> Decode.field("metric_tests", Decode.array(metricDecoder)),
        });
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
        specific: json |> testSpecificInfo(testType),
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