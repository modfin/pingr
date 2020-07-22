[@react.component]
let make =
    (
      ~label,
      ~errorMsg,
      ~infoText,
      ~pairs,
      ~onChange: Form.t => unit,
      ~keyPlaceholder,
      ~valuePlaceholder,
    ) => {
  let setPair = (n, currPairs, updatedPair) =>
    currPairs
    |> List.mapi((i, pair) =>
         if (i == n) {
           updatedPair;
         } else {
           pair;
         }
       );

  let rec drop = (n, index, list) => {
    switch (list) {
    | [] => []
    | [_, ...tail] when index == n => tail
    | [head, ...tail] => [head, ...drop(n, index + 1, tail)]
    };
  };

  let removePair = (n, currPairs) => drop(n, 0, currPairs);

  let addPair = currPairs => {
    currPairs @ [("", "")];
  };

  <div className="w-full px-3 mb-6 md:mb-0">
    <label
      className="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2">
      {label |> React.string}
    </label>
    <div className="flex flex-wrap -mx-3 mb-2">
      <div className="w-6/12 px-3">
        <label
          className="block uppercase tracking-wide text-gray-700 text-xs font-bold">
          {"Key" |> React.string}
        </label>
      </div>
      <div className="w-5/12 px-3">
        <label
          className="block uppercase tracking-wide text-gray-700 text-xs font-bold">
          {"Value" |> React.string}
        </label>
      </div>
    </div>
    {pairs
     |> List.mapi((i, (key, value)) =>
          <div key={string_of_int(i)} className="flex flex-wrap -mx-3 mb-3">
            <div className="w-6/12 px-3">
              <input
                value=key
                placeholder=keyPlaceholder
                onChange={e =>
                  Form.List(
                    setPair(
                      i,
                      pairs,
                      (ReactEvent.Form.target(e)##value, value),
                    ),
                  )
                  |> onChange
                }
                className="appearance-none block w-full bg-gray-200 text-gray-700 border border-gray-400 rounded py-3 px-4 leading-tight focus:outline-none focus:bg-white"
              />
              {i == List.length(pairs) - 1
                 ? errorMsg != React.null
                     ? <p className="text-red-500 text-xs italic">
                         errorMsg
                       </p>
                     : <p className="text-gray-600 text-xs italic">
                         {infoText |> React.string}
                       </p>
                 : React.null}
            </div>
            <div className="w-5/12 px-3">
              <input
                value
                placeholder=valuePlaceholder
                onChange={e =>
                  Form.List(
                    setPair(
                      i,
                      pairs,
                      (key, ReactEvent.Form.target(e)##value),
                    ),
                  )
                  |> onChange
                }
                className="appearance-none block w-full bg-gray-200 text-gray-700 border border-gray-400 rounded py-3 px-4 leading-tight focus:outline-none focus:bg-white"
              />
            </div>
            <div className="w-1/12 px-3">
              <button
                type_="button"
                onClick={_e => Form.List(removePair(i, pairs)) |> onChange}
                className="self-center bg-red-500 hover:bg-red-700 text-white py-1 px-2 rounded w-8 h-8">
                {"-" |> React.string}
              </button>
            </div>
          </div>
        )
     |> Array.of_list
     |> React.array}
    <button
      type_="button"
      onClick={_e => Form.List(addPair(pairs)) |> onChange}
      className="bg-green-500 hover:bg-green-700 text-white py-1 px-2 rounded w-8 h-8">
      {"+" |> React.string}
    </button>
  </div>;
};