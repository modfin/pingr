[@react.component]
let make =
    (
      ~label,
      ~errorMsg,
      ~infoText,
      ~list,
      ~onChange: Form.t => unit,
      ~placeholder,
    ) => {
  let setElem = (n, currElems, updatedElem) =>
    currElems
    |> List.mapi((i, elem) =>
         if (i == n) {
           updatedElem;
         } else {
           elem;
         }
       );

  let rec drop = (n, index, list) => {
    switch (list) {
    | [] => []
    | [_, ...tail] when index == n => tail
    | [head, ...tail] => [head, ...drop(n, index + 1, tail)]
    };
  };

  let removeElem = (n, currElems) => drop(n, 0, currElems);

  let addElem = currElems => {
    currElems @ [""];
  };

  <div className="w-full px-3 mb-6 md:mb-0">
    <label
      className="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2">
      {label |> React.string}
    </label>
    <div className="flex flex-wrap -mx-3 mb-2">
      <div className="w-full px-3">
        <label
          className="block uppercase tracking-wide text-gray-700 text-xs font-bold">
          {"Values" |> React.string}
        </label>
      </div>
    </div>
    {list
     |> List.mapi((i, value) =>
          <div key={string_of_int(i)} className="flex flex-wrap -mx-3 mb-3">
            <div className="w-11/12 px-3">
              <input
                value
                placeholder
                onChange={e =>
                  Form.List(
                    setElem(i, list, ReactEvent.Form.target(e)##value),
                  )
                  |> onChange
                }
                className="appearance-none block w-full bg-gray-200 text-gray-700 border border-gray-400 rounded py-3 px-4 leading-tight focus:outline-none focus:bg-white"
              />
              {i == List.length(list) - 1
                 ? errorMsg != React.null
                     ? <p className="text-red-500 text-xs italic">
                         errorMsg
                       </p>
                     : <p className="text-gray-600 text-xs italic">
                         {infoText |> React.string}
                       </p>
                 : React.null}
            </div>
            <div className="w-1/12 px-3">
              {i != 0
                 ? <button
                     type_="button"
                     onClick={_e =>
                       Form.List(removeElem(i, list)) |> onChange
                     }
                     className="self-center bg-red-500 hover:bg-red-700 text-white py-1 px-2 rounded w-8 h-8">
                     {"-" |> React.string}
                   </button>
                 : React.null}
            </div>
          </div>
        )
     |> Array.of_list
     |> React.array}
    <button
      type_="button"
      onClick={_e => Form.List(addElem(list)) |> onChange}
      className="bg-green-500 hover:bg-green-700 text-white py-1 px-2 rounded w-8 h-8">
      {"+" |> React.string}
    </button>
  </div>;
};