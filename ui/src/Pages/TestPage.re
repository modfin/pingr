type testTypes =
  | HTTP
  | Prometheus
  | TLS
  | DNS
  | Ping
  | SSH
  | TCP
  | HTTPPush
  | PrometheusPush;

module Loadable = {
  type t('result) =
    | Loading
    | Failed(string)
    | Success('result);
};

type action =
  | LoadData
  | LoadSuccess(Models.Test.t)
  | LoadFail(string);

type testState =
  | NotAsked
  | Loading
  | Success(Models.Test.t)
  | Failure;

[@react.component]
let make = (~id) => {
  let (state, dispatch) =
    React.useReducer(
      (_state, action) =>
        switch (action) {
        | LoadData => Loadable.Loading
        | LoadSuccess(posts) => Loadable.Success(posts)
        | LoadFail(msg) => Loadable.Failed(msg)
        },
      Loadable.Loading,
    );

  React.useEffect0(() => {
    Api.fetchTestWithCallback(id, result =>
      switch (result) {
      | None => dispatch(LoadFail("Not working"))
      | Some(test) => dispatch(LoadSuccess(test))
      }
    );
    None;
  });

  let (errorMsg, setErrorMsg) = React.useState(() => "");

  let deleteCallback = resp => {
    switch (resp) {
    | Api.Error(msg) => setErrorMsg(_ => msg)
    | Api.Success(_)
    | Api.SuccessJSON(_) => Paths.goToTests()
    };
  };

  switch (state) {
  | Loading => <div> {ReasonReact.string("Loading...")} </div>
  | Failed(msg) => <div> {ReasonReact.string(msg)} </div>
  | Success(test) =>
    <>
      <div className="relative bg-gray-400 my-4 p-1">
        <p className="text-xl font-bold ml-1">
          {"Test status: " ++ test.testName |> React.string}
        </p>
        <button
          onClick={_e => Paths.goToEditTest(id)}
          className="m-1 bg-blue-500 hover:bg-blue-700 text-white py-1 px-2 rounded">
          {"Edit" |> React.string}
        </button>
        <button
          onClick={_e => Api.deleteTest(test.testId, deleteCallback)}
          className="m-1 bg-red-500 hover:bg-red-700 text-white py-1 px-2 rounded">
          {"Delete" |> React.string}
        </button>
        {errorMsg != ""
           ? <p className="text-red-500 text-xs italic">
               {"Error deleteing test: " ++ errorMsg |> React.string}
             </p>
           : React.null}
      </div>
      <div className="grid grid-rows-2 grid-cols-2  gap-4 p-2">
        <div className="col-span-2 row-span-1 md:col-span-1 md:row-span-2">
          <p className="text-lg font-bold"> {"Standard" |> React.string} </p>
          <DataField labelName="Type" value={test.testType} />
          <DataField labelName="Url" value={test.url} />
          <DataField
            labelName="Interval"
            value={string_of_int(test.interval) ++ " s"}
          />
          <DataField
            labelName="Timeout"
            value={string_of_int(test.timeout) ++ " s"}
          />
          <DataField
            labelName="Created"
            value={Js.Date.toLocaleString(
              Js.Date.fromString(test.createdAt),
            )}
          />
        </div>
        <div className="col-span-2 row-span-1 md:col-span-1 md:row-span-2">
          <p className="text-lg font-bold"> {"Specific" |> React.string} </p>
          {switch (test.specific) {
           | TLS(port)
           | TCP(port) => <DataField labelName="Port" value={port.port} />
           | HTTP(http) =>
             <>
               <DataField labelName="Method" value={http.method_} />
               <DataField
                 labelName="Payload"
                 value={
                   switch (http.payload) {
                   | None => "-"
                   | Some(payload) => payload
                   }
                 }
               />
               <DataField
                 labelName="Expected response"
                 value={
                   switch (http.expResult) {
                   | None => "-"
                   | Some(expResult) => expResult
                   }
                 }
               />
             </>
           | Prometheus(promMetrics) =>
             Models.Test.(
               {
                 promMetrics.metrics
                 |> Array.mapi((i, pMetric) => {
                      <div key={pMetric.key}>
                        <DataField
                          labelName={"Key " ++ string_of_int(i)}
                          value={pMetric.key}
                        />
                        <div className="ml-10 md:mr-5">
                          <DataField
                            labelName="Lower bound"
                            value={string_of_float(pMetric.lowerBound)}
                          />
                          <DataField
                            labelName="Upper bound"
                            value={string_of_float(pMetric.upperBound)}
                          />
                          <DataField
                            labelName="Labels"
                            value={
                              switch (pMetric.labels) {
                              | None => "-"
                              | Some(labels) =>
                                let keys = Js.Dict.keys(labels);
                                let values = Js.Dict.values(labels);
                                let dictString = ref("{ ");
                                for (i in 0 to Array.length(keys) - 1) {
                                  dictString :=
                                    dictString^
                                    ++ keys[i]
                                    ++ ": "
                                    ++ values[i]
                                    ++ ", ";
                                };
                                dictString := dictString^ ++ "}";
                                dictString^;
                              }
                            }
                          />
                        </div>
                      </div>
                    })
                 |> React.array;
               }
             )
           | Empty => <DataField labelName="None" value="-" />
           }}
        </div>
      </div>
      <TestContactsList testId={test.testId} />
      <TestUptime testId={test.testId} />
      <LogList testId={test.testId} />
    </>
  };
};