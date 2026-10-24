# fsrs_reference.json

`fsrs_reference.json` holds 400 simulated cards (`cards`: per card a list of
`[rating, delta_t, time]`) and what fsrs-rs 2.0.4 (FSRS-5, the model go-fsrs
v3 implements) computes from them: `compute_parameters` with short-term
scheduling on (`params`), and `evaluate` with the default and the fitted
parameters (log loss and RMSE). `TestOptimizeMatchesFsrsRs` compares the Go
optimizer with it.

To make it again:

1. `FSRS_FIXTURE_OUT=/some/dir/fixture go test ./scheduler -run TestWriteReferenceFixture`
   writes `fixture.cards.json` and `fixture.items.txt` (the training items in
   the order the Go optimizer uses, one per line: `rating,delta_t ...`).
2. Copy the `fsrs-2.0.4` crate sources, add the example below as
   `examples/reference.rs` (with an `[[example]]` entry in `Cargo.toml`), and
   build it with the crate's own `Cargo.lock` (some of its dependencies are
   yanked): `cargo build --release --example reference`.
3. Run `target/release/examples/reference fixture.items.txt` and put its
   output, with the cards, into `fsrs_reference.json`.

```rust
// Reads FSRS items (one per line: "rating,delta_t rating,delta_t ...") from
// the file given as the first argument and prints the parameters
// compute_parameters fits (short-term on), and the log loss and RMSE (bins)
// of the default and the fitted parameters.
use fsrs::{FSRSItem, FSRSReview, FSRS};

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let path = std::env::args().nth(1).expect("usage: reference <items file>");
    let text = std::fs::read_to_string(path)?;
    let items: Vec<FSRSItem> = text
        .lines()
        .filter(|l| !l.trim().is_empty())
        .map(|l| FSRSItem {
            reviews: l
                .split_whitespace()
                .map(|p| {
                    let (r, t) = p.split_once(',').unwrap();
                    FSRSReview { rating: r.parse().unwrap(), delta_t: t.parse().unwrap() }
                })
                .collect(),
        })
        .collect();
    let fsrs = FSRS::new(Some(&[]))?;
    let params = fsrs.compute_parameters(items.clone(), None, true)?;
    println!("params {:?}", params);
    let d = fsrs.evaluate(items.clone(), |_| true)?;
    println!("default {} {}", d.log_loss, d.rmse_bins);
    let o = FSRS::new(Some(&params))?.evaluate(items, |_| true)?;
    println!("fitted {} {}", o.log_loss, o.rmse_bins);
    Ok(())
}
```
