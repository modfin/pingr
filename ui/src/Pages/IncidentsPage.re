module Loadable = {
  type t('result) =
    | Loading
    | Failed(string)
    | Success('result);
};

type action =
  | LoadData
  | LoadSuccess(list(Models.Incident.t))
  | LoadFail(string);

[@react.component]
let make = () => {
  let (page, setPage) = React.useState(_ => 0);

  let incidentsPerPage = 10;

  let (state, dispatch) =
    React.useReducer(
      (_state, action) =>
        switch (action) {
        | LoadData => Loadable.Loading
        | LoadSuccess(incidents) => Loadable.Success(incidents)
        | LoadFail(msg) => Loadable.Failed(msg)
        },
      Loadable.Loading,
    );

  React.useEffect1(
    () => {
      if (state == Loadable.Loading) {
        Api.fetchIncidentsWithCallback(result =>
          switch (result) {
          | None =>
            dispatch(
              LoadFail(
                "Not working, perhaps you haven't had any incidents yet?",
              ),
            )
          | Some(incidents) => dispatch(LoadSuccess(incidents))
          }
        );
      };
      None;
    },
    [|state|],
  );

  Models.Incident.(
    <>
      <Divider title="Incidents" />
      <div className="py-2 divide-y-4 divide-gray-400">
        {switch (state) {
         | Loading =>
           <div className="px-6"> {"Loading..." |> React.string} </div>
         | Failed(msg) => <div className="px-6"> {msg |> React.string} </div>
         | Success(incidents) =>
           <>
             <div className="px-6">
               <p className="text-lg font-bold">
                 {"Active" |> React.string}
               </p>
               <table className="w-full text-left mb-2">
                 {let active =
                    incidents |> List.filter(incident => incident.active);
                  List.length(active) != 0
                    ? <>
                        <thead className="flex w-full">
                          <tr className="flex w-full">
                            <td
                              className="font-bold px-4 py-2 lg:w-1/12 w-1/6">
                              {"Test name" |> React.string}
                            </td>
                            <td
                              className="font-bold px-4 py-2 lg:w-1/12 w-1/6">
                              {"Details" |> React.string}
                            </td>
                            <td className="font-bold px-4 py-2 lg:w-2/3 w-1/2">
                              {"Root cause" |> React.string}
                            </td>
                            <td className="font-bold px-4 py-2 w-1/6">
                              {"Created" |> React.string}
                            </td>
                          </tr>
                        </thead>
                        <tbody className="mb-4">
                          {active
                           |> List.map(incident => {
                                <tr
                                  key={string_of_int(incident.incidentId)}
                                  className="w-full flex">
                                  <td
                                    className="border px-4 py-2 lg:w-1/12 w-1/6">
                                    <a
                                      className="no-underline text-blue-500 hover:underline cursor-pointer"
                                      onClick={_event =>
                                        Paths.goToTest(incident.testId)
                                      }>
                                      {incident.testName |> React.string}
                                    </a>
                                  </td>
                                  <td
                                    className="border px-4 py-2 lg:w-1/12 w-1/6">
                                    <a
                                      className="no-underline text-blue-500 hover:underline cursor-pointer"
                                      onClick={_event =>
                                        Paths.goToIncident(
                                          string_of_int(incident.incidentId),
                                        )
                                      }>
                                      {"View" |> React.string}
                                    </a>
                                  </td>
                                  <td
                                    className="border px-4 py-2 lg:w-2/3 w-1/2">
                                    {incident.rootCause |> React.string}
                                  </td>
                                  <td className="border px-4 py-2 w-1/6">
                                    {incident.createdAt
                                     |> Js.Date.fromString
                                     |> Js.Date.toLocaleString
                                     |> React.string}
                                  </td>
                                </tr>
                              })
                           |> Array.of_list
                           |> React.array}
                        </tbody>
                      </>
                    : <tbody>
                        <tr>
                          <td className=" py-2 italic">
                            {"No active incidents" |> React.string}
                          </td>
                        </tr>
                      </tbody>}
               </table>
             </div>
             <div className="px-6">
               <p className="text-lg font-bold mt-2">
                 {"Closed" |> React.string}
               </p>
               <table className="w-full text-left">
                 {let closed =
                    incidents
                    |> List.filter(incident => !incident.active)
                    |> Array.of_list;
                  Array.length(closed) != 0
                    ? <>
                        <thead className="flex w-full">
                          <tr className="flex w-full">
                            <td
                              className="font-bold px-4 py-2 lg:w-1/12 w-1/6">
                              {"Test name" |> React.string}
                            </td>
                            <td
                              className="font-bold px-4 py-2 lg:w-1/12 w-1/6">
                              {"Details" |> React.string}
                            </td>
                            <td className="font-bold px-4 py-2 lg:w-1/2 w-1/3">
                              {"Root cause" |> React.string}
                            </td>
                            <td className="font-bold px-4 py-2 w-1/6">
                              {"Created" |> React.string}
                            </td>
                            <td className="font-bold px-4 py-2 w-1/6">
                              {"Closed" |> React.string}
                            </td>
                          </tr>
                        </thead>
                        <tbody>
                          {closed
                           |> Belt.Array.slice(
                                ~offset=page * incidentsPerPage,
                                ~len=incidentsPerPage,
                              )
                           |> Array.map(incident => {
                                <tr
                                  key={string_of_int(incident.incidentId)}
                                  className="w-full flex">
                                  <td
                                    className="border px-4 py-2 w-1/6 lg:w-1/12">
                                    <a
                                      className="no-underline text-blue-500 hover:underline cursor-pointer"
                                      onClick={_event =>
                                        Paths.goToTest(incident.testId)
                                      }>
                                      {incident.testName |> React.string}
                                    </a>
                                  </td>
                                  <td
                                    className="border px-4 py-2 w-1/6 lg:w-1/12">
                                    <a
                                      className="no-underline text-blue-500 hover:underline cursor-pointer"
                                      onClick={_event =>
                                        Paths.goToIncident(
                                          string_of_int(incident.incidentId),
                                        )
                                      }>
                                      {"View" |> React.string}
                                    </a>
                                  </td>
                                  <td
                                    className="border px-4 py-2 lg:w-1/2 w-1/3">
                                    {incident.rootCause |> React.string}
                                  </td>
                                  <td className="border px-4 py-2 w-1/6">
                                    {incident.createdAt
                                     |> Js.Date.fromString
                                     |> Js.Date.toLocaleString
                                     |> React.string}
                                  </td>
                                  <td className="border px-4 py-2 w-1/6">
                                    {incident.closedAt
                                     |> Js.Date.fromString
                                     |> Js.Date.toLocaleString
                                     |> React.string}
                                  </td>
                                </tr>
                              })
                           |> React.array}
                        </tbody>
                      </>
                    : <tr>
                        <td className="py-2 italic">
                          {"No closed incidents" |> React.string}
                        </td>
                      </tr>}
               </table>
               {List.length(incidents) > incidentsPerPage
                  ? <ul className="flex list-reset mt-1">
                      <li>
                        <a
                          className="block hover:text-white hover:bg-blue-700 cursor-pointer border px-3 py-2"
                          onClick={_ =>
                            setPage(curr =>
                              if (curr > 0) {
                                curr - 1;
                              } else {
                                curr;
                              }
                            )
                          }>
                          {"Previous" |> React.string}
                        </a>
                      </li>
                      <li className="ml-1">
                        <a
                          className="block hover:text-white hover:bg-blue-700 cursor-pointer border px-3 py-2"
                          onClick={_ =>
                            setPage(curr =>
                              if (curr
                                  < List.length(
                                      incidents |> List.filter(i => !i.active),
                                    )
                                  / incidentsPerPage) {
                                curr + 1;
                              } else {
                                curr;
                              }
                            )
                          }>
                          {"Next" |> React.string}
                        </a>
                      </li>
                    </ul>
                  : React.null}
             </div>
           </>
         }}
      </div>
    </>
  );
};