[@react.component]
let make = (~id) => {
  let (fullIncident, setFullIncident) = React.useState(_ => None);

  React.useEffect0(() => {
    let cb = result => {
      setFullIncident(_ => result);
    };
    Api.fetchIncidentWithCallback(id, cb);
    None;
  });

  Models.IncidentFull.(
    <>
      <Divider title="Incident" />
      <div className="py-2 px-6">
        {switch (fullIncident) {
         | None => "Loading..." |> React.string
         | Some(fullIncident) =>
           <>
             <div className="table w-full lg:w-1/2">
               <div className="table-row-group">
                 <div className="table-row">
                   <div className="table-cell font-bold">
                     {"Test name:" |> React.string}
                   </div>
                   <div className="table-cell">
                     <a
                       className="no-underline text-blue-500 hover:underline cursor-pointer"
                       onClick={_event =>
                         Paths.goToTest(fullIncident.incident.testId)
                       }>
                       {fullIncident.incident.testName |> React.string}
                     </a>
                   </div>
                 </div>
                 <div className="table-row">
                   <div className="table-cell font-bold">
                     {"Created:" |> React.string}
                   </div>
                   <div className="table-cell">
                     {fullIncident.incident.createdAt
                      |> Js.Date.fromString
                      |> Js.Date.toLocaleString
                      |> React.string}
                   </div>
                 </div>
                 <div className="table-row">
                   <div className="table-cell font-bold">
                     {"Active:" |> React.string}
                   </div>
                   <div className="table-cell">
                     {fullIncident.incident.active
                      |> string_of_bool
                      |> React.string}
                   </div>
                 </div>
                 <div className="table-row">
                   <div className="table-cell font-bold">
                     {"Root cause:" |> React.string}
                   </div>
                   <div className="table-cell">
                     {fullIncident.incident.rootCause |> React.string}
                   </div>
                 </div>
                 <div className="table-row">
                   <div className="table-cell font-bold">
                     {{
                        fullIncident.incident.active
                          ? "Duration (ongoing):" : "Duration:";
                      }
                      |> React.string}
                   </div>
                   <div className="table-cell">
                     {let compareTime =
                        fullIncident.incident.active
                          ? Js.Date.now()
                          : Js.Date.getTime(
                              Js.Date.fromString(
                                fullIncident.incident.closedAt,
                              ),
                            );

                      let duration =
                        Js.Date.fromFloat(
                          compareTime
                          -. Js.Date.getTime(
                               Js.Date.fromString(
                                 fullIncident.incident.createdAt,
                               ),
                             ),
                        );

                      let hours =
                        Js.Date.getHours(duration) -. 1. |> Js.Float.toString;
                      let minutes =
                        duration |> Js.Date.getMinutes |> Js.Float.toString;
                      let seconds =
                        duration |> Js.Date.getSeconds |> Js.Float.toString;
                      {
                        {
                          hours == "0" ? "" : hours ++ " hour(s) ";
                        }
                        ++ {
                          minutes == "0" ? "" : minutes ++ " min(s) ";
                        }
                        ++ seconds
                        ++ " second(s)";
                      }
                      |> React.string}
                   </div>
                 </div>
                 <br />
                 <div className="table-row">
                   <div className="table-cell font-bold">
                     {"Contact log:" |> React.string}
                   </div>
                 </div>
               </div>
             </div>
             {switch (fullIncident.contactLog) {
              | Some(contactLog) =>
                Models.IncidentContactLog.(
                  <table className="table-auto text-left">
                    <thead>
                      <tr>
                        <th className="px-4 py-2">
                          {"Contact name" |> React.string}
                        </th>
                        <th className="px-4 py-2">
                          {"Message sent" |> React.string}
                        </th>
                        <th className="px-4 py-2">
                          {"Sent" |> React.string}
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      {contactLog
                       |> List.map(log => {
                            <tr key={log.contactId}>
                              <td className="border px-4 py-2">
                                {log.contactName |> React.string}
                              </td>
                              <td className="border px-4 py-2">
                                {log.message |> React.string}
                              </td>
                              <td className="border px-4 py-2">
                                {log.createdAt
                                 |> Js.Date.fromString
                                 |> Js.Date.toLocaleString
                                 |> React.string}
                              </td>
                            </tr>
                          })
                       |> Array.of_list
                       |> React.array}
                    </tbody>
                  </table>
                )
              | None => "None contacted" |> React.string
              }}
           </>
         }}
      </div>
    </>
  );
};