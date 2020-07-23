let first = list => List.length(list) > 0 ? Some(List.hd(list)) : None;

let getError = (field, errors) =>
  List.filter(((name, _)) => name === field, errors)
  |> first
  |> (
    errors =>
      switch (errors) {
      | Some((_, msgs)) => List.hd(msgs) |> React.string
      | None => ReasonReact.null
      }
  );

let setJsonKey = (dict, keyName, value) => {
  switch (value) {
  | Form.Str(s) => Js.Dict.set(dict, keyName, Js.Json.string(s))
  | Form.Float(f) => Js.Dict.set(dict, keyName, Js.Json.number(f))
  | Form.Int(i) =>
    Js.Dict.set(dict, keyName, Js.Json.number(float_of_int(i)))
  | Form.List(l) =>
    Js.Dict.set(
      dict,
      keyName,
      Js.Json.array(
        l |> List.map(item => Js.Json.string(item)) |> Array.of_list,
      ),
    )
  | Form.TupleList(l) =>
    let parsedJson = Js.Dict.empty();
    l
    |> List.iter(((key, value)) => {
         Js.Dict.set(parsedJson, key, Js.Json.string(value))
       });
    Js.Dict.set(dict, keyName, Js.Json.object_(parsedJson));
  | Form.PromMetrics(metrics) =>
    let metricsArr =
      metrics
      |> List.map(metric =>
           Models.Test.(
             {
               let metricJson = Js.Dict.empty();
               Js.Dict.set(metricJson, "key", Js.Json.string(metric.key));
               Js.Dict.set(
                 metricJson,
                 "lower_bound",
                 Js.Json.number(metric.lowerBound),
               );
               Js.Dict.set(
                 metricJson,
                 "upper_bound",
                 Js.Json.number(metric.upperBound),
               );
               let labelsJson = Js.Dict.empty();
               metric.labels
               |> List.iter(label =>
                    Js.Dict.set(
                      labelsJson,
                      fst(label),
                      Js.Json.string(snd(label)),
                    )
                  );
               Js.Dict.set(
                 metricJson,
                 "labels",
                 Js.Json.object_(labelsJson),
               );
               metricJson;
             }
           )
         )
      |> Array.of_list;
    Js.Dict.set(dict, keyName, Js.Json.objectArray(metricsArr));
  };
};
let emptyMsg = "Field is required";
let aboveZero = "Field has to be > 0";
let thresholdMsg = "Threshold has to be > 0";

let validContactThreshold = (value, _) => {
  switch (value) {
  | Form.TupleList(l) =>
    switch (
      List.find(testContact => int_of_string(snd(testContact)) <= 0, l)
    ) {
    | exception Not_found => true
    | _testContact => false
    }
  | _ => true
  };
};

let getContactsPayload = (testId, contacts_) => {
  let payload =
    switch (contacts_) {
    | Form.TupleList(contacts) =>
      contacts
      |> List.map(((contactId, threshold)) => {
           let contact = Js.Dict.empty();
           setJsonKey(contact, "contact_id", Str(contactId));
           setJsonKey(contact, "test_id", Str(testId));
           setJsonKey(contact, "threshold", Int(int_of_string(threshold)));
           contact;
         })
    | _ => []
    };
  payload |> Array.of_list;
};

let keyPairValidation = (value, _) => {
  switch (value) {
  | Form.TupleList(l) =>
    switch (List.find(((key, value)) => key == "" || value == "", l)) {
    | exception Not_found => true
    | _testContact => false
    }

  | _ => true
  };
};

let listValidation = (value, _) => {
  switch (value) {
  | Form.List(l) =>
    switch (List.find(value => value == "", l)) {
    | exception Not_found => true
    | _elem => false
    }

  | _ => true
  };
};