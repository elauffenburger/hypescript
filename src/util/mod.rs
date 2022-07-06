use std::{cell::RefCell, rc::Rc};

pub fn rcref<T>(t: T) -> Rc<RefCell<T>> {
    Rc::new(RefCell::new(t))
}
