type t =
  | Str(string)
  | Float(float)
  | Int(int);

type validation('a) =
  | NotEmpty
  | Custom('a);

let validate = (rule, value, values) =>
  switch (rule, value) {
  | (NotEmpty, Str(s)) => s != ""
  | (NotEmpty, Int(i)) => i > 0
  | (NotEmpty, Float(f)) => f > 0.
  | (Custom(fn), _) => fn(value, values)
  };

let validateField = (validations, value, values) =>
  List.fold_left(
    (errors, (fn, msg)) =>
      validate(fn, value, values) ? errors : List.append(errors, [msg]),
    [],
    validations,
  );

let validation = (rules, get, data) =>
  List.fold_left(
    (errors, (field, validations)) =>
      {
        let value = get(field, data);
        validateField(validations, value, data);
      }
      |> (
        fieldErrors =>
          List.length(fieldErrors) > 0
            ? List.append(errors, [(field, fieldErrors)]) : errors
      ),
    [],
    rules,
  );

module type Config = {
  type state;
  type field;
  let update: (field, t, state) => state;
  let get: (field, state) => t;
};

module FormComponent = (Config: Config) => {
  type field = Config.field;
  type values = Config.state;
  type errors = list((field, list(string)));
  type state = {
    errors,
    values,
  };
  type form = {
    form: state,
    handleChange: (field, t) => unit,
  };

  type action =
    | UpdateValues(field, t);

  [@react.component]
  let make = (~initialState, ~rules, ~render) => {
    let (state, dispatch) =
      React.useReducer(
        (state, action) =>
          switch (action) {
          | UpdateValues(field, value) =>
            let values = Config.update(field, value, state.values);
            {values, errors: validation(rules, Config.get, values)};
          },
        {
          values: initialState,
          errors: validation(rules, Config.get, initialState),
        },
      );
    let handleChange = (field, value) => {
      dispatch(UpdateValues(field, value));
    };

    render({form: state, handleChange});
  };
};