[@react.component]
let make = (~onChange, ~errorMsg, ~metrics) => {
  let emptyMetric =
    Models.Test.{key: "", lowerBound: 0., upperBound: 1., labels: []};

  let setMetric = (n, metrics, newMetric) => {
    (
      switch (metrics) {
      | Form.PromMetrics(m) => m
      | _ => []
      }
    )
    |> List.mapi((i, metric) =>
         if (i == n) {
           newMetric;
         } else {
           metric;
         }
       );
  };

  let addMetric = metrics => {
    switch (metrics) {
    | Form.PromMetrics(m) => m @ [emptyMetric]
    | _ => []
    };
  };

  let rec drop = (n, index, list) => {
    switch (list) {
    | [] => []
    | [_, ...tail] when index == n => tail
    | [head, ...tail] => [head, ...drop(n, index + 1, tail)]
    };
  };

  let removeMetric = (n, metrics) => {
    (
      switch (metrics) {
      | Form.PromMetrics(m) => m
      | _ => []
      }
    )
    |> drop(n, 0);
  };

  <div className="w-full mb-6 md:mb-0">
    <div className="w-full px-3 mb-6 md:mb-0">
      <label
        className="block -mx-3 uppercase tracking-wide text-gray-700 text-xs font-bold mb-2">
        {"Metrics" |> React.string}
      </label>
    </div>
    {Models.Test.(
       (
         switch (metrics) {
         | Form.PromMetrics(m) => m
         | _ => []
         }
       )
       |> List.mapi((i, metric) => {
            <div
              key={string_of_int(i)}
              className={
                "border border-gray-600 p-2 mb-1"
                ++ {
                  i != 0 ? " mt-5" : "";
                }
              }>
              <div className="flex flex-wrap px-3 -mx-3 mb-3">
                <label
                  className="block uppercase tracking-wide text-gray-700 text-sm font-bold">
                  {"("
                   ++ string_of_int(i + 1)
                   ++ ")"
                   ++ " Prometheus Key"
                   ++ {
                     i == 0 ? " (*)" : "";
                   }
                   |> React.string}
                </label>
                {i != 0
                   ? <button
                       type_="button"
                       onClick={_e => onChange(removeMetric(i, metrics))}
                       className="self-center bg-red-500 hover:bg-red-700 text-white py-1 px-2 rounded text-xs ml-2">
                       {"Remove" |> React.string}
                     </button>
                   : React.null}
              </div>
              <div className="flex flex-wrap px-3 -mx-3 mb-3">
                <input
                  value={metric.key}
                  onChange={e => {
                    let key = ReactEvent.Form.target(e)##value;
                    onChange(
                      setMetric(
                        i,
                        metrics,
                        {
                          key,
                          lowerBound: metric.lowerBound,
                          upperBound: metric.upperBound,
                          labels: metric.labels,
                        },
                      ),
                    );
                  }}
                  placeholder="scylla_cache_bytes_used"
                  className="appearance-none block w-full bg-gray-200 text-gray-700 border border-gray-400 rounded py-3 px-4 leading-tight focus:outline-none focus:bg-white"
                />
                <p className="text-gray-600 text-xs italic">
                  {"The Prometheus key which will be checked" |> React.string}
                </p>
              </div>
              <div className="flex flex-wrap -mx-3 mb-6">
                <FormKeyValue
                  pairs={metric.labels}
                  label="prometheus key labels"
                  infoText="Labels in addition to the promethues key"
                  keyPlaceholder="type"
                  valuePlaceholder="gauge"
                  errorMsg=React.null
                  onChange={v =>
                    switch (v) {
                    | TupleList(labels) =>
                      onChange(
                        setMetric(
                          i,
                          metrics,
                          Models.Test.{
                            key: metric.key,
                            lowerBound: metric.lowerBound,
                            upperBound: metric.upperBound,
                            labels,
                          },
                        ),
                      )

                    | _ => ()
                    }
                  }
                />
              </div>
              <div className="flex flex-wrap px-3 -mx-3 mb-2">
                <div className="w-6/12">
                  <label
                    className="block uppercase tracking-wide text-gray-700 text-xs font-bold">
                    {"Lower bound" |> React.string}
                  </label>
                </div>
                <div className="w-6/12 pl-3">
                  <label
                    className="block uppercase tracking-wide text-gray-700 text-xs font-bold">
                    {"Upper bound" |> React.string}
                  </label>
                </div>
              </div>
              <div className="flex flex-wrap px-3 -mx-3 mb-3">
                <div className="w-6/12 pr-3">
                  <input
                    value={Js.Float.toString(metric.lowerBound)}
                    onChange={e => {
                      let v = ReactEvent.Form.target(e)##value;
                      let f =
                        try(Js.Float.fromString(v)) {
                        | _ => 0.
                        };
                      onChange(
                        setMetric(
                          i,
                          metrics,
                          Models.Test.{
                            key: metric.key,
                            lowerBound: f,
                            upperBound: metric.upperBound,
                            labels: metric.labels,
                          },
                        ),
                      );
                    }}
                    type_="number"
                    placeholder="0."
                    className="appearance-none block w-full bg-gray-200 text-gray-700 border border-gray-400 rounded py-3 px-4 leading-tight focus:outline-none focus:bg-white"
                  />
                  <p className="text-gray-600 text-xs italic">
                    {"Lower bound of absolute GAUGE value or increase of COUNTER value"
                     |> React.string}
                  </p>
                </div>
                <div className="w-6/12 pl-3">
                  <input
                    value={Js.Float.toString(metric.upperBound)}
                    onChange={e => {
                      let v = ReactEvent.Form.target(e)##value;
                      let f =
                        try(Js.Float.fromString(v)) {
                        | _ => 0.
                        };
                      onChange(
                        setMetric(
                          i,
                          metrics,
                          Models.Test.{
                            key: metric.key,
                            lowerBound: metric.lowerBound,
                            upperBound: f,
                            labels: metric.labels,
                          },
                        ),
                      );
                    }}
                    type_="number"
                    placeholder="10."
                    className="appearance-none block w-full bg-gray-200 text-gray-700 border border-gray-400 rounded py-3 px-4 leading-tight focus:outline-none focus:bg-white"
                  />
                  <p className="text-gray-600 text-xs italic">
                    {"Upper bound of absolute GAUGE value or increase of COUNTER value"
                     |> React.string}
                  </p>
                </div>
              </div>
            </div>
          })
     )
     |> Array.of_list
     |> React.array}
    {errorMsg != React.null
       ? <p className="text-red-500 text-xs italic"> errorMsg </p>
       : <p className="text-gray-600 text-xs italic">
           {"The prometheus key(s) which values(s) will be checked"
            |> React.string}
         </p>}
    <button
      type_="button"
      onClick={_e => onChange(addMetric(metrics))}
      className="bg-green-500 hover:bg-green-700 text-white py-1 px-2 rounded mb-6 mt-1">
      {"Add another metric" |> React.string}
    </button>
  </div>;
};