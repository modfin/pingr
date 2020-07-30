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
  let (showData, setShowData) = React.useState(_ => false);
  let (errorMsg, setErrorMsg) = React.useState(() => "");

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

  React.useEffect1(
    () => {
      if (state == Loadable.Loading) {
        Api.fetchTestWithCallback(id, result =>
          switch (result) {
          | None => dispatch(LoadFail("Not working"))
          | Some(test) => dispatch(LoadSuccess(test))
          }
        );
      };
      None;
    },
    [|state|],
  );

  let handleActiveChange = value => {
    Api.updateActive(id, value, result => {
      switch (result) {
      | Api.Error(msg) => setErrorMsg(_ => msg)
      | Api.Success(_)
      | Api.SuccessJSON(_) => dispatch(LoadData)
      }
    });
  };

  let deleteCallback = resp => {
    switch (resp) {
    | Api.Error(msg) => setErrorMsg(_ => msg)
    | Api.Success(_)
    | Api.SuccessJSON(_) => Paths.goToTests()
    };
  };

  let stringOfTupleList = tupleList => {
    let dictString = ref("{ ");
    tupleList
    |> List.iteri((i, tuple) => {
         dictString := dictString^ ++ fst(tuple) ++ ": " ++ snd(tuple);
         if (i != List.length(tupleList) - 1) {
           dictString := dictString^ ++ ", ";
         };
       });
    dictString := dictString^ ++ " }";
    dictString^;
  };

  let stringOfDict = dict => {
    let l = Js.Dict.entries(dict) |> Array.to_list;
    stringOfTupleList(l);
  };

  switch (state) {
  | Loading => <div> {ReasonReact.string("Loading...")} </div>
  | Failed(msg) => <div> {ReasonReact.string(msg)} </div>
  | Success(test) =>
    <>
      <Divider title={"Test: " ++ test.testName}>
        <div className="flex">
          <b> {"Status: " |> React.string} </b>
          <div className="mt-1 ml-1">
            <TestStatusLabel
              active={test.active}
              statusId={
                switch (test.statusId) {
                | None => 0
                | Some(i) => i
                }
              }
            />
          </div>
        </div>
        <p>
          {<b> {"Url: " |> React.string} </b>}
          {(
             switch (test.testType) {
             | "HTTPPush"
             | "PrometheusPush" =>
               "/api/push/"
               ++ test.testId
               ++ "/"
               ++ String.trim(test.testName)
             | _ => test.url
             }
           )
           |> React.string}
        </p>
        <button
          type_="button"
          onClick={_e => handleActiveChange(!test.active)}
          className="my-1 bg-gray-500 hover:bg-gray-700 text-white py-1 px-2 rounded">
          {(test.active ? "Pause" : "Activate") |> React.string}
        </button>
        <button
          type_="button"
          onClick={_e => Paths.goToEditTest(id)}
          className="ml-2 bg-blue-500 hover:bg-blue-700 text-white py-1 px-2 rounded">
          {"Edit" |> React.string}
        </button>
        <button
          type_="button"
          onClick={_e => Api.deleteTest(test.testId, deleteCallback)}
          className="ml-2 bg-red-500 hover:bg-red-700 text-white py-1 px-2 rounded">
          {"Delete" |> React.string}
        </button>
        <button
          type_="button"
          onClick={_e => setShowData(prev => !prev)}
          className="float-right my-1 bg-blue-500 hover:bg-blue-700 text-white py-1 px-2 rounded">
          {{
             showData ? "Hide data" : "Show data";
           }
           |> React.string}
        </button>
        {errorMsg != ""
           ? <p className="text-red-500 text-xs italic">
               {"Error deleteing test: " ++ errorMsg |> React.string}
             </p>
           : React.null}
      </Divider>
      {showData
         ? <>
             <div className="grid grid-rows-2 grid-cols-2  gap-4 px-6 py-2">
               <div
                 className=" col-span-2 row-span-1 md:col-span-1 md:row-span-2">
                 <p className="text-lg font-bold">
                   {"Standard" |> React.string}
                 </p>
                 <DataField labelName="Type" value={test.testType} />
                 {switch (test.testType) {
                  | "HTTPPush"
                  | "PrometheusPush" => React.null
                  | _ =>
                    <>
                      <DataField labelName="Url" value={test.url} />
                      <DataField
                        labelName="Interval"
                        value={string_of_int(test.interval) ++ " s"}
                      />
                    </>
                  }}
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
               <div
                 className="col-span-2 row-span-1 md:col-span-1 md:row-span-2">
                 <p className="text-lg font-bold">
                   {"Specific" |> React.string}
                 </p>
                 {switch (test.specific) {
                  | TLS(port)
                  | TCP(port) =>
                    <DataField labelName="Port" value={port.port} />
                  | HTTP(http) =>
                    <>
                      <DataField
                        labelName="Request Method"
                        value={http.reqMethod}
                      />
                      {switch (http.reqHeaders) {
                       | None => React.null
                       | Some(headers) =>
                         headers != Js.Dict.empty()
                           ? <DataField
                               labelName="Request headers"
                               value={stringOfDict(headers)}
                             />
                           : React.null
                       }}
                      {switch (http.reqBody) {
                       | None => React.null
                       | Some(body) =>
                         body != ""
                           ? <DataField labelName="Request body" value=body />
                           : React.null
                       }}
                      {switch (http.resStatus) {
                       | None => React.null
                       | Some(status) =>
                         status != 0
                           ? <DataField
                               labelName="Response status"
                               value={string_of_int(status)}
                             />
                           : React.null
                       }}
                      {switch (http.resHeaders) {
                       | None => React.null
                       | Some(headers) =>
                         headers != Js.Dict.empty()
                           ? <DataField
                               labelName="Response headers"
                               value={stringOfDict(headers)}
                             />
                           : React.null
                       }}
                      {switch (http.resBody) {
                       | None => React.null
                       | Some(body) =>
                         body != ""
                           ? <DataField labelName="Request body" value=body />
                           : React.null
                       }}
                    </>
                  | Prometheus(promMetrics) =>
                    Models.Test.(
                      {
                        promMetrics.metrics
                        |> List.mapi((i, pMetric) => {
                             <div key={pMetric.key}>
                               <DataField
                                 labelName={"Key " ++ string_of_int(i)}
                                 value={pMetric.key}
                               />
                               <div className="ml-10 md:mr-5">
                                 <DataField
                                   labelName="Lower bound"
                                   value={Js.Float.toString(
                                     pMetric.lowerBound,
                                   )}
                                 />
                                 <DataField
                                   labelName="Upper bound"
                                   value={Js.Float.toString(
                                     pMetric.upperBound,
                                   )}
                                 />
                                 <DataField
                                   labelName="Labels"
                                   value={stringOfTupleList(pMetric.labels)}
                                 />
                               </div>
                             </div>
                           })
                        |> Array.of_list
                        |> React.array;
                      }
                    )
                  | DNS(dns) =>
                    Models.Test.(
                      <div>
                        <DataField labelName="Record" value={dns.record} />
                        <DataField labelName="Strategy" value={dns.strategy} />
                        <DataField
                          labelName="Check"
                          value={List.nth(dns.check, 0)}
                        />
                      </div>
                    )
                  | SSH(ssh) =>
                    Models.Test.(
                      <div>
                        <DataField labelName="Port" value={ssh.port} />
                        <DataField labelName="Username" value={ssh.username} />
                        <DataField
                          labelName="Credential type"
                          value={ssh.credentialType}
                        />
                      </div>
                    )
                  | Empty => <DataField labelName="None" value="-" />
                  }}
               </div>
             </div>
             <TestContactsList testId={test.testId} />
           </>
         : <div className="h-8">
             <p className="text-center mt-2 italic">
               {"Test data hidden" |> React.string}
             </p>
           </div>}
      <TestStatus id={test.testId} />
    </>
  };
};